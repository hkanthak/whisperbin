FROM debian:bookworm-slim

WORKDIR /app
COPY whisperbin .

EXPOSE 80
ENTRYPOINT ["./whisperbin"]
