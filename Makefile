GOROOT := $(shell go env GOROOT)
DIST := dist

.PHONY: test build web serve clean

test:
	go test ./lib/engine ./lib/ai

build:
	go build -o ultimate-tic-tac-toe ./cmd/main

# The game page plus the AI worker, which runs in one Web Worker per CPU core.
web:
	mkdir -p $(DIST)
	GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o $(DIST)/main.wasm ./cmd/main
	GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o $(DIST)/aiworker.wasm ./cmd/aiworker
	install -m 644 "$(GOROOT)/lib/wasm/wasm_exec.js" wasm/index.html wasm/ai-pool.js wasm/ai-worker.js $(DIST)/

serve: web
	python3 -m http.server --directory $(DIST) 8080

clean:
	rm -rf $(DIST) ultimate-tic-tac-toe
