# MVP Roadmap

## Scope (MVP)

- File-based HTTP request definitions (`.req.yaml`)
- Env and secret interpolation
- HTTP execution with pretty response output
- Assertion engine (status, headers, JSONPath)
- Local run history + replay + diff
- Basic WebSocket connect/send/listen + transcript persistence

## Out of Scope (MVP)

- OAuth interactive flows
- GraphQL introspection UI
- Team cloud sync
- Full TUI app

## Milestones

1. **Foundation**
- CLI skeleton and config loading
- request parser + schema validation
- env resolution and interpolation

2. **HTTP Execution**
- request execution engine
- response renderer
- assertion engine

3. **History**
- run persistence (`.wirepad/history/`)
- replay by run id
- response diff

4. **WebSocket**
- connect/send/listen commands
- frame formatting and filtering
- transcript save/replay

5. **Interop**
- import from curl and HTTPie
- minimal Postman collection import

## Suggested Repo Layout

```text
cmd/wirepad/
internal/cli/
internal/config/
internal/requestspec/
internal/httpclient/
internal/wsclient/
internal/assert/
internal/history/
internal/render/
requests/
env/
docs/
.wirepad/
  env/
  history/
```

## First 10 Implementation Tickets

1. Bootstrap CLI (`req`, `send`, `hist`, `diff`, `replay`, `ws`) command tree.
2. Define and enforce `.req.yaml` schema for `kind=http|ws`.
3. Implement request resolver (`name` -> `requests/**`) with clear errors.
4. Implement variable interpolation with layered precedence.
5. Build HTTP sender with timeout, query, headers, and body modes.
6. Add human-friendly response renderer with optional `--json`.
7. Implement assertion engine for status/header/JSONPath checks.
8. Persist each run to `.wirepad/history/runs/<run_id>.json`.
9. Implement `replay` and `diff --last`.
10. Implement WS connect/send/listen and transcript save as NDJSON.

## Success Criteria

- Creating/editing/sending a request takes under 30 seconds for first-time setup.
- A failed assertion shows exact failing path and operator.
- Replay and diff work with zero manual response file handling.
- WS sessions can be captured and replayed from transcript files.
