.PHONY: build install test fmt vet tidy check clean assets docs check-docs check-integrations

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS ?= -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

build:
	go build -ldflags "$(LDFLAGS)" -o bin/rdrop ./cmd/rdrop

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/rdrop

test:
	go test ./...

fmt:
	gofmt -w ./cmd ./internal ./scripts

vet:
	go vet ./...

tidy:
	go mod tidy

check: fmt tidy vet test build check-docs

check-docs: build
	./bin/rdrop help --markdown | diff - docs/COMMANDS.md

docs: build
	./bin/rdrop help --markdown > docs/COMMANDS.md

check-integrations:
	node --test packaging/npm/launcher.test.cjs
	npm --prefix integrations/executor run check
	npm --prefix integrations/executor test
	python3 -m unittest discover -s integrations/laya -p '*_test.py'

assets:
	go run ./scripts/render-terminal-demo.go > docs/assets/terminal-demo.svg

clean:
	rm -rf bin dist
