FROM golang:1.25-bookworm AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/whisperbin ./cmd/whisperbin

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=build /out/whisperbin .
COPY ui ./ui

EXPOSE 8080
ENTRYPOINT ["./whisperbin"]
