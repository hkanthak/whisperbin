# -------- Build Stage --------
FROM golang:1.24 AS builder

# Arbeitsverzeichnis im Build-Container
WORKDIR /app

# Moduldateien zuerst kopieren und Abhängigkeiten cachen
COPY go.mod go.sum ./
RUN go mod download

# Restlichen Source-Code kopieren
COPY . .

RUN go build -ldflags="-s -w" -o whisperbin

# -------- Runtime Stage --------
FROM debian:bookworm-slim

# Nur das fertige Binary kopieren
WORKDIR /app
COPY --from=builder /app/whisperbin .

# Webserver Port
EXPOSE 80

# Startbefehl
ENTRYPOINT ["./whisperbin"]
