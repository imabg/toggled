# Feature 02 — Flag Management

**Depends on:** [`01-project-foundation.md`](01-project-foundation.md)
**Shared context:** see [`../plan.md`](../plan.md) for architecture, data
model summary, and API conventions.

## Goal

CRUD API for creating and managing feature flags: boolean, string, or
arbitrary-JSON (`object`) values, each with an `enabled` state and a
`default_value` used when disabled.

## Data model

Migration adds the `flags` table:

```sql
CREATE TYPE flag_type AS ENUM ('boolean', 'string', 'json');

CREATE TABLE flags (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    key            text NOT NULL UNIQUE,
    name           text NOT NULL,
    description    text NOT NULL DEFAULT '',
    type           flag_type NOT NULL,
    enabled        boolean NOT NULL DEFAULT false,
    value          jsonb NOT NULL,
    default_value  jsonb NOT NULL,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_flags_key ON flags (key);
```

`key` format: lowercase slug, `^[a-z0-9]+(-[a-z0-9]+)*$` (e.g.
`new-checkout-flow`), max 100 chars. Enforce in the service layer (a DB
`CHECK` constraint is optional defense-in-depth).

Type/value matching (`type=boolean` → JSON boolean, `type=string` → JSON
string, `type=json` → any valid JSON value) is validated in the **service
layer** before every insert/update. A DB `CHECK` using `jsonb_typeof` may be
added as a second line of defense but is not a substitute for application
validation (it can't distinguish "any JSON" from a specific shape).

## Domain type

```go
type FlagType string

const (
    FlagTypeBoolean FlagType = "boolean"
    FlagTypeString  FlagType = "string"
    FlagTypeJSON    FlagType = "json"
)

type Flag struct {
    ID            uuid.UUID
    Key           string
    Name          string
    Description   string
    Type          FlagType
    Enabled       bool
    Value         json.RawMessage
    DefaultValue  json.RawMessage
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

## Endpoints (`/api/v1/flags`)

### `POST /api/v1/flags` — create

Request:

```json
{
  "key": "new-checkout-flow",
  "name": "New checkout flow",
  "description": "Enables the redesigned checkout experience",
  "type": "boolean",
  "enabled": false,
  "value": true,
  "default_value": false
}
```

Response `201`: the created flag (same shape, plus `id`, `created_at`,
`updated_at`).

Errors:
- `400 invalid_value` — `value`/`default_value` doesn't match `type`.
- `400 invalid_key` — key fails the slug format.
- `409 flag_exists` — `key` already in use.

### `GET /api/v1/flags` — list

Response `200`:

```json
{ "flags": [ { "id": "...", "key": "new-checkout-flow", "...": "..." } ] }
```

v1 returns all flags unpaginated (flag counts are expected to be small).
Note pagination (`limit`/`offset` or cursor) as a documented future
enhancement, not required now.

### `GET /api/v1/flags/{key}` — get one

`200` with the flag, or `404 flag_not_found`.

### `PATCH /api/v1/flags/{key}` — update

Request body may include any subset of `name`, `description`, `enabled`,
`value`, `default_value`. **`key` and `type` are immutable** — reject
attempts to change them with `400 immutable_field` (to change a flag's
type, delete and recreate it). Re-validate `value`/`default_value` against
the existing `type` on update.

`200` with the updated flag, or `404 flag_not_found`.

### `DELETE /api/v1/flags/{key}` — delete

`204` on success, `404 flag_not_found` if missing.

## Error envelope

Follow the convention in [`../plan.md`](../plan.md#api-conventions):

```json
{ "error": { "code": "flag_not_found", "message": "no flag exists with key \"x\"" } }
```

## Out of scope

- Authentication/authorization (authx, upstream of this service).
- Environments, projects, or any multi-tenancy concept.
- User targeting rules, segments, percentage rollout.

## Acceptance criteria

- Full CRUD works for all three types (`boolean`, `string`, `json`),
  including a nested-object example for `json`.
- Creating a flag with mismatched `type`/`value` returns `400`.
- Creating a flag with a duplicate `key` returns `409`.
- Attempting to change `key` or `type` via `PATCH` returns `400`.
- Deleting, then fetching the same key returns `404`.
- Unit tests cover the service-layer validation (type/value matching, key
  format) independent of HTTP; integration tests cover each endpoint
  against a real Postgres instance (`testcontainers-go`).
