.PHONY: build test run docker docker-run clean

## Build the binary for the current OS/arch (for local development)
build:
	go build -o findingway .

## Build a Linux amd64 binary (matches Docker target)
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o findingway .

## Build a MacOS arm64 binary (matches development machine)
build-macos:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s" -o findingway .

## Run tests with race detector
test:
	go test ./... -race -covermode=atomic -coverprofile=coverage.out

## Run locally (requires DISCORD_TOKEN in environment)
run: build
	./findingway

## Build the Docker image
docker:
	docker build -t findingway .

## Run via Docker Compose
docker-run:
	docker compose up --build

## Remove build artifacts
clean:
	rm -f findingway findingway.exe coverage.out
