# CLAUDE.md

Guidance for working on WhisperBin with Claude Code.

## What this is

One-time secret sharing web app in Go. Secrets are stored encrypted in memory (no database, no logs) and deleted after a single retrieval or TTL expiry. An optional "secure mode" requires out-of-band passcode approval before the secret is delivered over SSE.

## Architecture

- `cmd/whisperbin/main.go` — entrypoint, HTTP server, periodic cleanup goroutine.
- `internal/constants.go` — tunables (TTL bounds, rate limits, lockout).
- `internal/storage/` — in-memory store (`store.go`), AES-GCM crypto (`crypto.go`), `Secret` type (`types.go`). A mutex-protected map keyed by a random 128-bit ID.
- `internal/web/` — HTTP handlers, routing, CSRF, rate limiting. One file per concern (`form_`, `recipient_`, `confirm_`, `sse_`, `status_`).
- `ui/templates/` — server-side rendered HTML (Pico.css). `ui/static/` — assets.

## Request flow

`GET /` (form) → `POST /secret` (store) → link `/{id}`.

- Non-secure: `GET /{id}` decrypts, deletes, shows once.
- Secure: recipient opens `/{id}` and connects to `GET /sse?id=`; the page shows a code the recipient relays to the sender out-of-band; the sender approves via `POST /confirm/{id}`; the secret then streams over SSE and is deleted.

## Conventions

- **No comments in code.** Keep code self-explanatory; do not add inline or `//` comments.
- Code must stay clean under `gofmt`, `go vet`, and `staticcheck`, and pass `go test ./... -race`.
- Keep web handlers thin; all secret/crypto logic lives in `internal/storage`.

## Commands

```
go run ./cmd/whisperbin        # serves on :8080
go build ./...
go test ./... -race
gofmt -l . && go vet ./... && staticcheck ./... && gosec ./...
```

## Configuration (env)

- `SECRET_KEY` — optional 32-byte base64 AES key (otherwise a random key per start). Storage is in-memory, so nothing survives a restart regardless of this value.
- `ALLOWED_ORIGIN` — allowed Origin for SSE and the base used to build the share link (default `http://localhost:8080`).
- `TRUST_PROXY` — set to `true` when running behind a reverse proxy so rate limiting and the confirm lockout key on the real client IP (`X-Forwarded-For` / `X-Real-IP`) instead of the proxy IP.

## Deployment

Runs behind any TLS-terminating reverse proxy (Traefik / Nginx / Caddy); a `Dockerfile` and an ARM CI workflow exist. Behind a proxy, set `TRUST_PROXY=true`. For private-only access on a self-hosted PaaS, restrict the route to a VPN source range via a proxy IP-allowlist middleware.

## Status & known issues

Addressed on branch `review/fixes`:

- Goroutine leak in secure-mode delivery: `WaitForUnlock` now takes a `context.Context` and selects on cancellation and TTL, and resets the listener flag so a dropped SSE connection can reconnect.
- The TTL form value is now parsed and clamped (it was previously ignored, always defaulting to 10 minutes).
- The HTTP server now sets read/idle timeouts (and intentionally no `WriteTimeout`, to keep long-lived SSE connections alive).
- Client IP behind a reverse proxy via `TRUST_PROXY`.

Still open (low severity):

- `gosec` G104: several `ExecuteTemplate` return values are unhandled — consider logging them.
- `gosec` G705: taint warnings on the SSE error line and the status JSON; low real risk (controlled strings, and the secret body is HTML-escaped). Revisit if the templating changes.
- One-time links can be consumed by URL preview scanners; use secure mode for sensitive shares.

## Learnings

- The trickiest part of the codebase is the secure-mode handoff (`WaitingCh` + `listenerSet` in `store.go`). Any change there must account for context cancellation, TTL expiry, and listener reconnect — block only inside a `select`, never on a bare channel receive.
- Run the full lint suite (`gofmt`, `vet`, `staticcheck`, `gosec`) before committing; `gosec` is worthwhile here because this is a security tool.
