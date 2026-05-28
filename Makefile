.PHONY: build run test lint clean docker kind-setup demo fmt tidy all

BINARY=polaris
BUILD_DIR=./build

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/polaris

run: build
	$(BUILD_DIR)/$(BINARY) serve --dry-run

test:
	go test -v -race -coverprofile=coverage.out ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BUILD_DIR) ./data

docker:
	docker build -t polaris:latest -f deployments/Dockerfile .

kind-setup:
	./scripts/setup-kind.sh

demo: build
	$(BUILD_DIR)/$(BINARY) serve --dry-run

fmt:
	go fmt ./...

tidy:
	go mod tidy

all: tidy fmt build test
