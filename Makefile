APP=tick

.PHONY: run build install tidy test

run:
	go run ./cmd/tick

build:
	mkdir -p bin
	go build -o bin/$(APP) ./cmd/tick

install:
	go install ./cmd/tick

tidy:
	go mod tidy

test:
	go test ./...