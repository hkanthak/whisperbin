FROM debian:bookworm-slim

WORKDIR /app

COPY whisperbin .

COPY ui/templates/ ./ui/templates/

EXPOSE 80
ENTRYPOINT ["./whisperbin"]
