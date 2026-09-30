run:
	go run ./cmd/server

test:
	go test -race ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./... 2>/dev/null || go vet ./...

build:
	go build -o bin/server ./cmd/server
