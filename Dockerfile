# ── Stage 1: build ────────────────────────────────────────────────────────────
FROM golang:1.22.1-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o findingway .

# ── Stage 2: runtime ──────────────────────────────────────────────────────────
FROM alpine:3.19.1

LABEL org.opencontainers.image.title="Findingway"
LABEL org.opencontainers.image.description="FFXIV Party Finder Discord bot"

WORKDIR /findingway

# Create directories for persistent data and logs
RUN mkdir -p /findingway/data /findingway/logs

COPY --from=builder /src/findingway .
COPY config.yaml .

# Persist the SQLite DB and log files on the host
VOLUME ["/findingway/data", "/findingway/logs"]

ENV DB_PATH=/findingway/data/findingway.db
ENV LOG_PATH=/findingway/logs/findingway.log
ENV CONFIG_PATH=/findingway/config.yaml

# DISCORD_TOKEN must be supplied at runtime via -e or docker-compose
ENTRYPOINT ["/findingway/findingway"]
