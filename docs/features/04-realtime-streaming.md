# Feature 04 — Real-Time Streaming

**Depends on:** [`02-flag-management.md`](02-flag-management.md)
**Shared context:** see [`../plan.md`](../plan.md) for the architecture
diagram and API conventions. Can be built in parallel with
[`03-flag-evaluation.md`](03-flag-evaluation.md) — they don't depend on each
other.

## Goal

Push flag changes to connected clients the moment they happen, instead of
requiring clients to poll. Implemented as a Server-Sent Events (SSE) stream
backed by Postgres `LISTEN`/`NOTIFY`, so it works correctly even if this
service later runs multiple replicas (no extra broker like Redis needed).

## Endpoint

### `GET /api/v1/stream`

- Response headers: `Content-Type: text/event-stream`,
  `Cache-Control: no-cache`, `Connection: keep-alive`.
- The connection is held open; the server writes one SSE event per flag
  change:

```
event: flag.updated
data: {"key":"new-checkout-flow","type":"boolean","enabled":true,"value":true}

```

- Event `event` field is one of `flag.created`, `flag.updated`,
  `flag.deleted`. `data` is a JSON payload with the flag's `key`, `type`,
  `enabled`, and evaluated `value` (deleted events omit `value`/`type`,
  just `{"key": "..."}`).
- Send a periodic keep-alive comment (e.g. `: ping\n\n`) every ~15–30s so
  intermediary proxies don't time out the idle connection.
- On client disconnect (context cancellation / write error), clean up the
  subscriber and stop writing to it.

## Mechanism

1. **Postgres side**: a trigger function on `flags` fires `AFTER INSERT OR
   UPDATE OR DELETE` and calls `pg_notify('flag_changed', payload)` where
   `payload` is a small JSON string (event type + key + minimal flag data;
   Postgres `NOTIFY` payloads are capped at 8000 bytes, so keep it small —
   don't dump large `json`-type flag values through the notify channel if
   they could be large; the listener can re-fetch full data if needed).
2. **Go listener** (`internal/stream`): one dedicated `pgx` connection
   issues `LISTEN flag_changed` and runs a loop calling
   `conn.WaitForNotification(ctx)`, parsing each notification and handing
   it to the in-process broadcaster. Reconnects with backoff if the
   listener connection drops.
3. **Broadcaster** (`internal/stream`): maintains a registry of connected
   SSE clients, each represented by a buffered channel. On a notification,
   fan the event out to every registered channel (non-blocking send — if a
   client's buffer is full, drop the oldest event or disconnect the slow
   client rather than blocking the broadcaster).
4. **HTTP handler**: on `GET /api/v1/stream`, registers a new channel with
   the broadcaster, flushes each incoming event to the `http.ResponseWriter`
   using `http.Flusher`, and unregisters on request context cancellation.

## Out of scope

- Authentication/authorization (authx, upstream of this service).
- Filtering the stream by flag key/subset (v1 streams all flag changes to
  every connected client).
- Guaranteed delivery / replay of missed events after a disconnect (clients
  should call `GET /api/v1/evaluate` to resync on reconnect).

## Acceptance criteria

- Connecting to `GET /api/v1/stream` and then creating/updating/deleting a
  flag via the management API produces a corresponding SSE event on the
  open connection within a small, bounded delay.
- Multiple concurrent stream clients all receive the same events.
- Disconnecting a client cleans up its broadcaster registration (no memory
  leak under repeated connect/disconnect).
- Keep-alive pings are sent on idle connections.
- Integration test: open a stream (e.g. via an SSE-aware test client),
  mutate a flag through the API, assert the event is received with the
  correct type and payload.
