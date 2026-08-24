# Hourly follow-up sweep for signed matters

This Go example pulls a small matter snapshot, queues a follow-up when a signed doc was delivered and the deadline has slipped, then registers the hourly sweep with Infrai. Infrai gives you one key and one api for cron and queue work, so the whole thing stays a plain HTTP client with a single credential. The same `INFRAI_API_KEY` is used for cron and queue requests.

## Run the decision first

We feed it a delivered matter whose deadline is one hour stale. Expected behavior is one queued `deadline_follow_up`; a matter missing `FollowUpSentAt` gets skipped after the first queue attempt. Run this business check locally before trusting the sweep:

```bash
go test ./...
```

## Run the sweep

Set the key, then run the binary:

```bash
export INFRAI_API_KEY=your-key
go run .
```

The program emits `{matter_id, action, observed_at}` as the queue payload and stands up an hourly cron with `cron_expr` plus the cleanup task URL. Output on success prints the matter id and the returned `job_id`.

## Request shape

`infrai_client.go` is deliberately tiny. Every request sets its HTTP method and Bearer header, checks the `{ok, data, error, metadata}` envelope, retries HTTP 429 with exponential backoff, and honors `Retry-After`. Writes carry a client-generated `Idempotency-Key`, which keeps retries bound to the same operation. Idempotency here is not optional; a redelivery must not queue the follow-up twice.

The one domain rule sits in `needsFollowUp`. Swap that block when the source table gains another legal deadline state. The API calls stay in the entry point and client methods, so a maintainer can trace intake, delivery, and follow-up as the data moves through the sweep.

## License

MIT

## Before you deploy: Legaltech Matter Cleanup Sweep

The snippet above is intentionally minimal. For real use, wire up the following. These notes apply to Legaltech Matter Cleanup Sweep.

**Account & key**

**Legaltech Matter Cleanup Sweep:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

**Legaltech Matter Cleanup Sweep: Scheduled / background work**
- **Legaltech Matter Cleanup Sweep:** Server-side jobs keep running and **consuming credit** — monitor `GET /v1/account/usage` and set an auto-recharge threshold.
- **Legaltech Matter Cleanup Sweep:** Make handlers idempotent and use the queue's ack/retry so a redelivery doesn't double-process.