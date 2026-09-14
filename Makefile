BINARY      := parameters-ldap
CMD         := ./cmd/server
IMAGE       := ghcr.io/jasonmiller-cc/parameters-ldap
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GOFLAGS     := -trimpath -ldflags="-s -w -X main.version=$(VERSION)"

.PHONY: all build test lint vet fmt tidy docker-build clean

all: build

build:
	go build $(GOFLAGS) -o bin/$(BINARY) $(CMD)

test:
	go test -race -count=1 ./...

lint:
	golangci-lint run ./...

vet:
	go vet ./...

fmt:
	gofmt -l -w .

tidy:
	go mod tidy

docker-build:
	docker buildx build \
		--build-context parameters-core=../parameters-core \
		--build-arg VERSION=$(VERSION) \
		-t $(IMAGE):$(VERSION) -t $(IMAGE):latest \
		--load .

clean:
	rm -rf bin/

run: build
	./bin/$(BINARY)
