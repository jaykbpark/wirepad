# CLI UX

This doc defines command ergonomics for wirepad.

## Design Principles

- Commands should be composable and scriptable.
- Interactive workflows should still be editor-first, not TUI-first.
- Naming should be short and memorable.

## Core Commands

```bash
# Create request artifacts
wirepad req new users/create POST https://api.example.com/users
wirepad req new events/subscribe WS wss://api.example.com/events

# Edit request in configured editor
wirepad req edit users/create

# Execute request
wirepad send users/create --env dev

# History and replay
wirepad hist users/create
wirepad replay 2026-02-17T10-21-11Z_7f3c

# Compare against previous response
wirepad diff users/create --last

# WebSocket direct mode
wirepad ws connect wss://api.example.com/events --env dev
wirepad ws send @payloads/subscription.json
wirepad ws listen --timeout 10s
wirepad ws save-transcript transcripts/events-01.ndjson
```

## Command Meanings

- `wirepad req new`: create a new `*.req.yaml` template file.
- `wirepad req edit`: open that file in `$VISUAL` or `$EDITOR`, then validate after close.
- `wirepad send`: execute request, print response, run assertions, and persist run history.
- `wirepad hist`: list previous runs for a request.
- `wirepad diff`: compare latest run against previous run (status, headers, body).
- `wirepad replay`: rerun from a saved run record.
- `wirepad ws *`: WebSocket connect/send/listen/transcript operations.

## Request Resolution

Input can be:

- request name (`users/create`)
- direct path (`requests/users/create.req.yaml`)

Resolution order:

1. exact path match
2. `requests/**/<name>.req.yaml`

## Output Modes

- default: colored pretty output for humans
- `--json`: machine-readable JSON result envelope
- `--quiet`: print only essential response fields

## Send Result Envelope (`--json`)

```json
{
  "run_id": "2026-02-17T10-21-11Z_7f3c",
  "request": "users/create",
  "ok": true,
  "status": 201,
  "duration_ms": 242,
  "assertions": {
    "passed": 4,
    "failed": 0
  },
  "history_path": ".wirepad/history/runs/2026-02-17T10-21-11Z_7f3c.json"
}
```

## Editor Workflow

`wirepad req edit <name>` should:

1. resolve request file
2. create it if missing (from template)
3. open in `$VISUAL`, then `$EDITOR`
4. validate on save/exit and show actionable errors

## Error UX

Errors should include:

- what failed
- where it failed (file path and field)
- how to fix quickly

Example:

```text
Validation error in requests/users/create.req.yaml
field: request.body.mode
value: jsno
hint: expected one of [json, raw, file, form, multipart]
```
