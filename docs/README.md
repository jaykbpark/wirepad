# wirepad

Postman-level ergonomics in a terminal-native workflow.

## Why This Exists

Current CLI request tools are great for one-offs, but painful for iterative API work:

- Editing JSON inline is slow and error-prone.
- Request history is not structured as reusable artifacts.
- WebSocket workflows are usually bolted on or missing.
- Team sharing and versioning are weak compared to file-based workflows.

This project aims to solve that by combining:

- CLI execution speed
- `$EDITOR`-based request editing
- file-first request definitions tracked in git
- first-class HTTP and WebSocket support

## Product Direction

Core principle: **execution in CLI, authoring in editor, history in files**.

Key capabilities:

- `edit-before-send` request workflow
- environment + secrets layering (`dev/stage/prod`)
- response history and diffing
- replay and automation-friendly command surface
- WebSocket connect/send/listen/transcript workflow
- import from curl, HTTPie, and Postman collections

## Docs

- `docs/request-file-spec.md`: exact `.req.yaml` format for HTTP and WS requests
- `docs/cli-ux.md`: command UX, command meanings, and example workflows
- `docs/architecture.md`: exact file structure and runtime persistence model
- `docs/mvp-roadmap.md`: MVP scope, milestones, and first implementation tickets
- `docs/git-workflow.md`: atomic commit and PR naming convention
