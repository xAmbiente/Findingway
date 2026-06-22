FROM golang:1.22.1 AS builder
WORKDIR /src
COPY . /src
RUN make build

FROM alpine:3.19.1
WORKDIR /findingway
# Create a volume mount point for the SQLite database
VOLUME ["/findingway/data"]
COPY --from=builder /src/findingway .
COPY --from=builder /src/config.yaml .
ENV DB_PATH=/findingway/data/findingway.db
ENTRYPOINT ["/findingway/findingway"]
