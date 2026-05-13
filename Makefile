
run: build
	./build/woo

fmt:
	gofmt -s -w ./cmd

.phony: version
version ?= $(shell git describe --tags --always --dirty)

.phony: build
build: fmt
	mkdir -p build
	go build \
	  -o build/woo \
	  -ldflags "-X main.Version=$(version)" \
	  cmd/*.go

.phony: lint
lint: fmt
	golangci-lint run ./...

.phony: vuln
vuln:
	govulncheck ./...

install: build
	mkdir -p ~/.local/bin
	cp build/woo ~/.local/bin/.
