# Feature 03 — Flag Evaluation (Runtime Retrieval)

**Depends on:** [`02-flag-management.md`](02-flag-management.md)
**Shared context:** see [`../plan.md`](../plan.md) for architecture and API
conventions.

## Goal

The read-only API that other applications call at runtime to retrieve
current flag values. This is the primary consumption path for Toggled and
should be simple, fast, and cheap to poll.

## Endpoints (`/api/v1/evaluate`)

### `GET /api/v1/evaluate` — bulk fetch

Returns every flag's evaluated value in one call, keyed by `key`. This is
the endpoint client SDKs poll on an interval or use to warm a local cache.

Response `200`:

```json
{
  "flags": {
    "new-checkout-flow": { "type": "boolean", "enabled": true, "value": true },
    "welcome-message":   { "type": "string",  "enabled": true, "value": "Hello!" },
    "pricing-config":    { "type": "json",    "enabled": false, "value": { "tier": "default" } }
  }
}
```

`value` is always the **evaluated** value: `flag.value` if `enabled`,
otherwise `flag.default_value`. Callers should not need to know about the
enabled/disabled distinction — `value` is already resolved. `enabled` is
included for observability/debugging.

### `GET /api/v1/evaluate/{key}` — single flag fetch

Response `200`:

```json
{ "key": "new-checkout-flow", "type": "boolean", "enabled": true, "value": true }
```

`404 flag_not_found` if the key doesn't exist.

## Behavior

- Evaluation logic: `value = flag.enabled ? flag.value : flag.default_value`.
- No request body, no per-caller context (no user attributes) — v1 has no
  targeting rules, so evaluation is global, not per-request.
- This is the hot path. For v1, querying Postgres directly per request is
  fine given expected flag volumes. Document (but do not require for v1) an
  optional future optimization: an in-memory read cache in the service,
  invalidated via the same Postgres `NOTIFY` mechanism used by
  [`04-realtime-streaming.md`](04-realtime-streaming.md), to avoid a DB
  round-trip on every evaluation call.

## Out of scope

- Authentication/authorization (authx, upstream of this service).
- Per-user/context-aware targeting, segments, percentage rollout.
- Write operations — this feature is strictly read-only (writes belong to
  [`02-flag-management.md`](02-flag-management.md)).

## Acceptance criteria

- `GET /api/v1/evaluate` returns all flags with correctly resolved values
  for each type, including when `enabled=false` (returns `default_value`).
- `GET /api/v1/evaluate/{key}` returns the correct evaluated value for an
  existing key and `404` for an unknown key.
- Integration tests seed flags of all three types in both enabled and
  disabled states and assert the evaluated `value` in each case.
