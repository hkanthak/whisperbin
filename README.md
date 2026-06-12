

<p align="center">
  <img src="ui/static/title_small.png" alt="WhisperBin Banner">
</p>

<p align="center">
  <img src="https://img.shields.io/github/go-mod/go-version/hkanthak/whisperbin?color=4b97c4" alt="Go Version">
  <a href="https://github.com/hkanthak/whisperbin/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/hkanthak/whisperbin?label=License&color=4b97c4" alt="License">
  </a>
  <img src="https://img.shields.io/github/last-commit/hkanthak/whisperbin?label=Last%20Commit&color=4b97c4" alt="Last Commit">
</p>

## One-Time Secret Sharing Tool for Developers

WhisperBin is a secure web application for sharing short text snippets (such as tokens, SSH keys, or passwords) via links that can be viewed **exactly once**. After retrieval, secrets are deleted automatically.

It is built in Go with server-side rendered HTML templates and optimized for simplicity, privacy, and technical transparency.

---

## Why I built WhisperBin

As a Kotlin developer, I like exploring new programming languages. Go is well suited for building small, fast server applications, and this project was a good way to try it out.

WhisperBin is designed to be lightweight and easy to run, for example on a Raspberry Pi in a home network. I use it to share passwords and other secrets between different devices.

---

## Features

- One-time retrieval (secret deleted on first access)
- Optional secure mode with manual recipient approval (passcode flow)
- Optional TTL (secret expires automatically after a configurable time)
- Encrypted in-memory storage (AES-GCM with per-instance key)
- No database required
- No storage of sensitive logs
- CSRF protection on all forms
- Per-IP rate limiting
- Minimal and clean UI, no JS frameworks required

---

## Usage Flow

1. User submits a secret via `/` form.
2. WhisperBin generates a random ID and stores the encrypted secret in memory.
3. A one-time link is generated: `https://yourhost/{id}`
4. The recipient visits the link:
   - If secure mode is enabled: recipient requests approval with a passcode.
   - Otherwise: a reveal page is shown; clicking "Reveal Secret" shows the secret and immediately deletes it.
5. The link is invalid after first use or after TTL expiration.

---

## Tech Stack

- **Backend**: Go (net/http, crypto/rand, html/template)
- **Frontend**: HTML templates (SSR) with [Pico.css](https://picocss.com/) for minimal styling
- **Storage**: In-memory map (sync.Mutex protected)
- **Routing**:
  - `GET /` — Submit secret form
  - `POST /secret` — Store secret
  - `GET /{id}` — Reveal page (does not consume the secret)
  - `POST /{id}` — Reveal and delete the secret (CSRF-protected)
  - `POST /confirm/{id}` — Manual approval (secure mode)
  - `GET /status/{id}` — Status polling (secure mode)
  - `GET /sse?id={id}` — SSE delivery (secure mode)

---

## Security Design

- **Random IDs**: 128-bit, securely generated with `crypto/rand`
- **Encryption**: AES-GCM with per-instance 256-bit key
- **One-time access**: Secret is deleted after first view; revealing requires a POST, so link-preview scanners cannot consume it
- **TTL support**: Expired secrets are automatically purged
- **Secure mode**: Optional manual recipient approval via passcode + SSE unlock flow
- **Rate limiting**: Per-IP rate limiting implemented (golang.org/x/time/rate)
- **CSRF**: All forms protected with CSRF tokens
- **No sensitive logging**: No storage of secret content or access logs

---

## Running Locally

```bash
git clone https://github.com/hkanthak/whisperbin.git
cd whisperbin
go run ./cmd/whisperbin
```

Access via: [http://localhost:8080](http://localhost:8080)

---

## Build & Deploy

```bash
go build -o whisperbin ./cmd/whisperbin
```

### Docker

The included multi-stage `Dockerfile` builds from source — no prebuilt binary required:

```bash
docker build -t whisperbin .
docker run -p 8080:8080 whisperbin
```

WhisperBin runs behind any TLS-terminating reverse proxy (Traefik, Nginx, Caddy). It can also be deployed straight from this Git repository by any Docker-based PaaS (e.g. Dokploy) that builds the image itself. Behind a proxy, set `ALLOWED_ORIGIN` to the public URL and `TRUST_PROXY=true`.

---

## Configuration

| Environment Variable | Description                                                                                     |
| -------------------- | ----------------------------------------------------------------------------------------------- |
| `SECRET_KEY`         | Optional 32-byte base64-encoded encryption key. If unset, a random key is generated at startup. |
| `ALLOWED_ORIGIN`     | Allowed origin for SSE connections and the base for generated links. Default: `http://localhost:8080`.          |
| `TRUST_PROXY`        | Set to `true` behind a reverse proxy so rate limiting uses the real client IP (`X-Forwarded-For` / `X-Real-IP`). |

---

## Project Structure

```
.
├── cmd/whisperbin/main.go          # Main entrypoint
├── internal/storage/               # In-memory storage + encryption logic
├── internal/web/                   # HTTP handlers, templates, CSRF, rate limiting
├── ui/templates/                   # HTML templates
├── ui/static/                      # CSS, favicon, optional images
└── README.md
```

---

## Screenshot

![WhisperBin Screenshot](ui/static/screenshot.png)

---

## Limitations / Disclaimer

WhisperBin is a minimal proof-of-concept for safe sharing of secrets in technical contexts. It is provided **as-is** without warranty. Do not use it for highly sensitive production data without independent security review.

---

## License

MIT License — see [LICENSE](LICENSE) file.
