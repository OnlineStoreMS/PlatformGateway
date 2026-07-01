.PHONY: run build tidy

run:
	go run ./cmd/gateway -config configs/config.yaml

build:
	go build -o bin/platform-gateway ./cmd/gateway

tidy:
	go mod tidy
