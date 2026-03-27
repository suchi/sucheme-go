.PHONY: all build test bench cover clean wasm

all: build test

build:
	go build -o sucheme ./cmd/sucheme

test:
	go test ./... -v

bench:
	go test -bench=. -benchmem ./...

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -f sucheme coverage.out coverage.html
	rm -f web-ui/sucheme.wasm web-ui/wasm_exec.js

wasm:
	GOOS=js GOARCH=wasm go build -o web-ui/sucheme.wasm ./cmd/wasm
	cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" web-ui/
