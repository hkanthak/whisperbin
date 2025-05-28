# WhisperBin

A minimal, secure tool written in Go for one-time sharing of sensitive text snippets like tokens or SSH keys.

## Purpose

Users can submit short text snippets that can be accessed **only once**. After being viewed, the secret is automatically deleted. An optional expiration time (TTL) can also be set.

## Features

- One-time access per secret
- In-memory or Redis-based storage
- Server-side rendering via `html/template`
- No sensitive data logging
- Optional TTL (time-to-live)
- Ready for deployment (Docker, GitHub Actions)

## Quickstart

```bash
go run ./cmd/whisperbin
```

Then open [http://localhost:8080](http://localhost:8080)

## Project Structure

```bash
cmd/whisperbin        → application entry point (main.go)
internal/web          → HTTP handlers and HTML templates
```

This project is still under development.
