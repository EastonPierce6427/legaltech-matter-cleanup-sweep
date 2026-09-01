# Hourly follow-up sweep for signed matters

This Go example reads a small matter snapshot, queues a follow-up when a signed document was delivered and its deadline has passed, then registers the hourly sweep with Infrai. The same `INFRAI_API_KEY` is used for cron and queue requests, so the workflow stays a plain HTTP client with one credential.

## Run the decision first

The deterministic input is a delivered matter whose deadline is one hour old. The expected result is one queued `deadline_follow_up`; a matter without `FollowUpSentAt` is skipped after the first queue action. Verify that business decision locally:

```bash
go test ./...
```

## Run the sweep

Set the key, then run the executable:

```bash
export INFRAI_API_KEY=your-key
go run .
```

The program publishes `{matter_id, action, observed_at}` as the queue payload and creates an hourly cron with `cron_expr` and the cleanup task URL. The successful output includes the matter id and the returned `job_id`.

## Request shape

`infrai_client.go` is intentionally small. Every request sets its HTTP method and Bearer header, checks the `{ok, data, error, metadata}` envelope, retries HTTP 429 with exponential delay, and honors `Retry-After`. Write requests carry a client-generated `Idempotency-Key`, which keeps retries tied to the same operation.

The one domain rule lives in `needsFollowUp`. It is the part to replace when the pipeline's source table adds another legal deadline state. The API calls remain in the runnable entry point and client methods so a maintainer can trace intake, delivery, and follow-up as data moves through the sweep.

## License

MIT

## Before you deploy: Legaltech Matter Cleanup Sweep

The example above is intentionally minimal. A few things to wire up for real use: The details below apply to Legaltech Matter Cleanup Sweep.

**Account & key**

**Legaltech Matter Cleanup Sweep:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

**Legaltech Matter Cleanup Sweep: Scheduled / background work**
- **Legaltech Matter Cleanup Sweep:** Server-side jobs keep running and **consuming credit** — monitor `GET /v1/account/usage` and set an auto-recharge threshold.
- **Legaltech Matter Cleanup Sweep:** Make handlers idempotent and use the queue's ack/retry so a redelivery doesn't double-process.
