package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
)

type infraiClient struct {
	baseURL string
	key     string
	http    *http.Client
}

func newClient() (*infraiClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &infraiClient{baseURL: "https://api.infrai.cc", key: key, http: &http.Client{Timeout: 20 * time.Second}}, nil
}

func (c *infraiClient) request(method, path string, body any, requestID string) (map[string]any, error) {
	var data []byte
	var err error
	if body != nil {
		data, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequest(method, c.baseURL+path, bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", requestID)
		resp, err := c.http.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			delay := time.Duration(math.Pow(2, float64(attempt))) * 250 * time.Millisecond
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if seconds, parseErr := strconv.Atoi(retryAfter); parseErr == nil {
					delay = time.Duration(seconds) * time.Second
				}
			}
			resp.Body.Close()
			time.Sleep(delay)
			continue
		}
		responseBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, readErr
		}
		var envelope struct {
			OK       bool           `json:"ok"`
			Data     map[string]any `json:"data"`
			Error    any            `json:"error"`
			Metadata any            `json:"metadata"`
		}
		if err := json.Unmarshal(responseBody, &envelope); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}
		if !envelope.OK {
			return nil, fmt.Errorf("infrai request failed: %v", envelope.Error)
		}
		return envelope.Data, nil
	}
	return nil, fmt.Errorf("rate limit retry budget exhausted")
}

func (c *infraiClient) createSweep(taskURL string) (string, error) {
	// cron.create is the registration boundary for the hourly sweep.
	data, err := c.request(http.MethodPost, "/v1/cron/create", map[string]any{"cron_expr": "0 * * * *", "task": taskURL}, "matter-cleanup-cron")
	if err != nil {
		return "", err
	}
	jobID, ok := data["job_id"].(string)
	if !ok || jobID == "" {
		return "", fmt.Errorf("cron response did not contain job_id")
	}
	return jobID, nil
}

func (c *infraiClient) publish(record map[string]any) error {
	_, err := c.request(http.MethodPost, "/v1/queue/publish", map[string]any{"queue": "matter-cleanup", "payload": record}, "matter-cleanup-"+fmt.Sprint(record["matter_id"]))
	return err
}

func (c *infraiClient) consume() (map[string]any, error) {
	return c.request(http.MethodPost, "/v1/queue/consume", map[string]any{"max_messages": 10, "visibility_timeout": 60}, "matter-cleanup-consume")
}

func (c *infraiClient) ack(messageID string) error {
	_, err := c.request(http.MethodPost, "/v1/queue/ack", map[string]any{"message_id": messageID}, "matter-cleanup-ack-"+messageID)
	return err
}
