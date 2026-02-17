# Request File Spec (`.req.yaml`)

This document defines the canonical request artifact used by the CLI.

## Goals

- Human-editable in any text editor
- Stable in git diffs
- Expressive enough for HTTP + WebSocket workflows
- Explicit env interpolation and test assertions

## File Naming

- Extension: `.req.yaml`
- Suggested location: `requests/<group>/<name>.req.yaml`
- Example: `requests/users/create.req.yaml`

## Environment Files

wirepad supports a shared env layer and a private env layer:

- Shared env: `env/<name>.env` (safe defaults, can be committed)
- Private env: `.wirepad/env/<name>.env` (secrets, gitignored)

Example:

- `env/dev.env`: `base_url=https://dev.api.example.com`
- `.wirepad/env/dev.env`: `token=...`

## Top-Level Schema

```yaml
version: 1
kind: http # http | ws
name: users.create
description: Create a user in the core API
tags: [users, create]

request: {}
expect: {}
hooks: {}
```

## HTTP Request Schema

```yaml
version: 1
kind: http
name: users.create

request:
  method: POST
  url: "{{base_url}}/users"
  query:
    invite: "true"
  headers:
    Authorization: "Bearer {{token}}"
    Content-Type: application/json
    X-Request-Id: "{{uuid}}"
  body:
    mode: json # json | raw | file | form | multipart
    json:
      email: "alice@example.com"
      name: "Alice"
  timeout_ms: 15000
  follow_redirects: true

expect:
  status: [200, 201]
  headers:
    content-type: "application/json"
  body:
    jsonpath:
      "$.id":
        exists: true
      "$.email":
        equals: "alice@example.com"
  max_duration_ms: 1000

hooks:
  pre_send:
    - set: { now_iso: "{{timestamp_iso}}" }
  post_receive:
    - export:
        user_id: "$.id"
```

## WebSocket Request Schema

```yaml
version: 1
kind: ws
name: events.subscribe

request:
  url: "wss://api.example.com/events"
  headers:
    Authorization: "Bearer {{token}}"
  connect_timeout_ms: 10000
  ping_interval_ms: 25000
  messages:
    - type: json # json | text | file
      json:
        op: subscribe
        topic: users
    - type: file
      path: payloads/custom-frame.json

expect:
  receive:
    - within_ms: 3000
      jsonpath:
        "$.op":
          equals: "subscribed"
```

## Interpolation Rules

`{{var}}` resolution order:

1. CLI `--var key=value`
2. active private environment file (`.wirepad/env/dev.env`, `.wirepad/env/stage.env`, etc.)
3. active shared environment file (`env/dev.env`, `env/stage.env`, etc.)
4. project `.env`
5. generated variables (for example: `uuid`, `timestamp_iso`)

If unresolved and no default is provided, execution fails with a clear error.

## Body Modes

- `json`: serialized JSON object
- `raw`: raw string payload
- `file`: file contents as request body
- `form`: urlencoded key/value pairs
- `multipart`: multipart form data with file parts

Example `file`:

```yaml
body:
  mode: file
  path: payloads/create-user.json
  content_type: application/json
```

## Assertion Operators

Supported operators:

- `equals`
- `not_equals`
- `exists`
- `contains`
- `matches` (regex)
- `gt`, `gte`, `lt`, `lte` (numeric)

Example:

```yaml
expect:
  body:
    jsonpath:
      "$.count":
        gte: 1
      "$.email":
        matches: "^[^@]+@[^@]+$"
```

## Hook Actions (MVP)

- `set`: set runtime variable
- `export`: map response JSONPath to named variable

Non-MVP hooks can add script execution later if needed.

## Validation Rules

- `version`, `kind`, `name`, and `request` are required.
- `kind=http` requires `request.method` and `request.url`.
- `kind=ws` requires `request.url`.
- Unknown fields are warnings in MVP, errors in strict mode (`--strict`).
