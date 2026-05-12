.PHONY: build install test fmt vet tidy check clean

build:
	go build -o bin/rdrop ./cmd/rdrop

install:
	go install ./cmd/rdrop

test:
	go test ./...

fmt:
	gofmt -w ./cmd ./internal

vet:
	go vet ./...

tidy:
	go mod tidy

check: fmt tidy vet test build

clean:
	rm -rf bin dist
