.PHONY: build run test lint clean docker kind-setup demo

BINARY=polaris
BUILD_DIR=./build

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/polaris

run:
	go run ./cmd/polaris serve --config ./configs/polaris.yaml

run-dry:
	go run ./cmd/polaris serve --config ./configs/polaris.yaml --dry-run

test:
	go test -v -race -coverprofile=coverage.out ./...

test-unit:
	go test -v -race ./internal/... ./pkg/...

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BUILD_DIR) data/

docker:
	docker build -t polaris:latest -f deployments/Dockerfile .

kind-setup:
	./scripts/setup-kind.sh

demo:
	go run ./cmd/polaris serve --config ./configs/polaris.yaml --dry-run
