# wirepad Architecture

This document defines the concrete filesystem layout and runtime logic.

## Repository Structure

```text
wirepad/
  cmd/
    wirepad/
      main.go
  internal/
    cli/
      root.go
      req.go
      send.go
      hist.go
      diff.go
      replay.go
      ws.go
    config/
      load.go
      env.go
      interpolate.go
      redact.go
    requestspec/
      schema.go
      parse.go
      validate.go
      resolve.go
    httpclient/
      build.go
      execute.go
    wsclient/
      connect.go
      stream.go
      transcript.go
    assert/
      eval.go
      jsonpath.go
    history/
      store.go
      list.go
      diff.go
      replay.go
    render/
      response.go
      table.go
  requests/
    users/
      create.req.yaml
  env/
    dev.env
    stage.env
    prod.env
  .wirepad/
    env/
      dev.env
      stage.env
      prod.env
    history/
      runs/
      bodies/
      index/
    transcripts/
  docs/
```

## Runtime Persistence Model

All runtime artifacts live under `.wirepad/` so project source remains clean.

- `.wirepad/env/*.env`
  - private per-environment values, usually secrets
  - gitignored
- `.wirepad/history/runs/<run_id>.json`
  - single source of truth for each execution record
- `.wirepad/history/bodies/<run_id>.resp`
  - response body store (to keep run json small)
- `.wirepad/history/index/<request_name>.json`
  - quick index of run IDs by request
- `.wirepad/transcripts/*.ndjson`
  - WebSocket session transcripts

## Environment Variable Strategy

Variables are resolved in this order:

1. `--var key=value` (highest priority)
2. `.wirepad/env/<env>.env` (private secrets)
3. `env/<env>.env` (shared defaults)
4. `.env` (project-level defaults)
5. generated runtime variables (for example `uuid`, `timestamp_iso`)

Policy:

- Real secrets should go in `.wirepad/env/*.env`.
- `env/*.env` should contain non-sensitive values or placeholders.
- `env/*.env.example` can be committed for onboarding.

## Command Execution Logic

`wirepad send <request> --env <name>` pipeline:

1. Resolve request path (`name` or direct file path)
2. Parse YAML into typed request spec
3. Validate schema and required fields
4. Resolve interpolation variables from env layers
5. Build transport request (HTTP or WS)
6. Execute request with timeout and options
7. Evaluate assertions
8. Render output (human or `--json`)
9. Persist history record and body artifacts
10. Return exit code:
  - `0`: transport success + assertions passed
  - `1`: transport error or assertion failure

## History Record Shape

Example run record:

```json
{
  "run_id": "2026-02-17T15-40-22Z_7f3c",
  "request_name": "users/create",
  "request_path": "requests/users/create.req.yaml",
  "env": "dev",
  "started_at": "2026-02-17T15:40:22Z",
  "duration_ms": 242,
  "ok": true,
  "status": 201,
  "assertions": {
    "passed": 4,
    "failed": 0
  },
  "response_headers": {
    "content-type": "application/json"
  },
  "body_ref": ".wirepad/history/bodies/2026-02-17T15-40-22Z_7f3c.resp"
}
```

## Redaction Rules

When rendering output or persisting history metadata, redact values for keys that match:

- `authorization`
- `token`
- `api_key`
- `secret`
- `password`

Body redaction is best-effort in MVP and fully configurable post-MVP.
