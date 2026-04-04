# PRD: Reliable Partitioned Webhook Ingestion

## Problem Statement
Current synchronous webhook ingestion in Synapse is vulnerable to data loss during system failures and suffers from "Head-of-Line" blocking when processing high-latency normalization logic. Furthermore, the lack of serial processing guarantees for a single order entity leads to race conditions and "Lifecycle Corruption" (e.g., a "Delivered" status update being overwritten by a delayed "Shipped" status).

## Goals / Non-goals
### Goals
- **Durability**: 100% persistence of raw webhook payloads before acknowledging receipt (HTTP 202).
- **Concurrency**: Parallel processing of webhooks across multiple Go routines.
- **Entity Serialism**: Guarantee that all events for a specific `VendorOrderID` are processed sequentially.
- **Observability**: Provide clear visibility into processing lag, retries, and failures.
- **Recoverability**: Automatic recovery of "stuck" jobs and a manual DLQ for permanent failures.

### Non-goals
- **Auto-Scaling**: Horizontal worker scaling (K8s HPA) is out of scope; we will use static partitioning.
- **Message Broker Dependency**: No external brokers like Redis/Kafka for V1; use a DB-backed "Outbox" pattern.

## User Personas
- **Developer**: Needs a reliable way to add new WMS/DIS providers without worrying about concurrency bugs.
- **Ops/Admin**: Needs to monitor processing lag and manually replay failed webhooks.

## User Stories
- **US1**: As a WMS Vendor, I want a <100ms response time for my webhooks so my systems don't time out.
- **US2**: As a Synapse System, I want to store the raw payload immediately so I can recover from a crash.
- **US3**: As a Developer, I want to ensure that "Shipped" and "Delivered" updates for Order #123 are never processed in parallel to avoid data corruption.
- **US4**: As an Admin, I want to see a list of "Poison Pill" webhooks that have failed 3 times so I can fix the data and re-trigger them.

## Acceptance Criteria
- [ ] **Outbox Pattern**: HTTP handler returns `202 Accepted` only after successful DB write to `RawWebhooks`.
- [ ] **Fast Key Extraction**: Use `gjson` to extract `VendorOrderID` during ingest for partitioning.
- [ ] **Static Partitioning**: Implement `N` background Go routines where each processes `hash(VendorOrderID) % N`.
- [ ] **Idempotency**: Webhooks with duplicate `VendorWebhookID` (if provided) are ignored.
- [ ] **Graceful Shutdown**: All workers receive a `SIGTERM` signal and finish their current job before exiting.
- [ ] **Sweeper**: A cron job periodically resets jobs stuck in `PROCESSING` status for longer than 10 minutes.
- [ ] **DLQ State**: Hard-fail after 3 retries, moving the row to `FAILED` for manual intervention.

## Edge Cases
- **Missing Order ID**: If `gjson` fails to find an ID, move to a "System Default" partition or a dedicated "Unsorted" queue for manual triage.
- **Hot Partitions**: One large client flooding its specific partition. Workers must stay isolated so other partitions continue at full speed.
- **Out-of-Order Source Timestamps**: Normalizer must skip updates if the incoming `source_updated_at` is older than the current DB state.

## Failure Modes
- **DB Write Failure (Ingest)**: Return 503 Service Unavailable to trigger vendor retry.
- **Normalization Panic**: Recover in the Go routine, increment `retry_count`, and log the stack trace.
- **Context Timeout**: Worker must abort long-running normalizations and release the job.

## Scalability Considerations (Flush & Drain)
- Current design uses static partitioning. To change the worker count (`N`), we implement a **"Flush and Drain"** protocol: Stop ingest server -> Wait for queue depth = 0 -> Restart with new config.
- Future move to **Postgres** is assumed to handle 1k+ WPS throughput.

## Future Extensions
- **NATS/Kafka Integration**: Replacing the DB-polling pattern with a real message broker.
- **Virtual Sharding**: Moving to 256 virtual shards to avoid "Flush and Drain" during scale-ups.
- **Notification Hooks**: Alerting Slack/Email when DLQ threshold is reached.
