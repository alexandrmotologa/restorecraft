.PHONY: all build test clean fixtures ui

BINARY_NAME=restorecraft
VERSION=1.0.0

all: test build

ui:
	@echo "Building React 19 Recovery Cockpit..."
	cd web && npm install && npm run build

build: ui
	@echo "Compiling standalone binary..."
	go build -ldflags="-s -w -X 'main.Version=$(VERSION)'" -o bin/$(BINARY_NAME) ./cmd/restorecraft

test:
	@echo "Running test suite..."
	go test -v -race ./...

fixtures:
	@echo "Generating synthetic testbeds..."
	go run ./cmd/restorecraft fixtures generate --dir ./test-backups

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/ dist/ web/dist web/node_modules test-backups/ data/
