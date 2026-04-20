APP=tick

.PHONY: run build install tidy test

run:
	go run ./cmd/tick

build:
	mkdir -p bin
	go build -o bin/$(APP) ./cmd/tick

install:
	go install ./cmd/tick

test:
	go test -short ./...

lint:
	golangci-lint version && golangci-lint run --verbose  -E  misspell    

generate-mocks:
	@echo "Generating mocks..."
	mockgen -source internal/usecase/usecase.go -destination internal/usecase/mocks/mock_interfaces.go -package mocks . PortfolioRepository,InstrumentRepository,PositionRepository
	@echo "Mocks generated successfully"
