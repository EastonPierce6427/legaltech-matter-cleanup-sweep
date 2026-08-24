package main

import (
	"fmt"
	"time"
)

type matter struct {
	ID             string
	SignedAt       time.Time
	DeliveredAt    time.Time
	Deadline       time.Time
	FollowUpSentAt time.Time
}

func needsFollowUp(m matter, now time.Time) bool {
	return !m.SignedAt.IsZero() && !m.DeliveredAt.IsZero() && now.After(m.Deadline) && m.FollowUpSentAt.IsZero()
}

func main() {
	client, err := newClient()
	if err != nil {
		panic(err)
	}
	now := time.Now().UTC()
	matters := []matter{{ID: "matter-104", SignedAt: now.Add(-72 * time.Hour), DeliveredAt: now.Add(-48 * time.Hour), Deadline: now.Add(-2 * time.Hour)}}
	for _, item := range matters {
		if !needsFollowUp(item, now) {
			continue
		}
		record := map[string]any{"matter_id": item.ID, "action": "deadline_follow_up", "observed_at": now.Format(time.RFC3339)}
		if err := client.publish(record); err != nil {
			panic(err)
		}
		fmt.Println("queued follow-up:", item.ID)
	}
	jobID, err := client.createSweep("https://example.com/legaltech/cleanup")
	if err != nil {
		panic(err)
	}
	fmt.Println("scheduled hourly sweep:", jobID)
}
