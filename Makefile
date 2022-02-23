wasm:
	GOARCH=wasm GOOS=js go build -ldflags='-w -s' -tags js,wasm \
	       -o ./dist/wasm/types.wasm ./cmd/wasm/main.go
	cp "$(shell go env GOROOT)/misc/wasm/wasm_exec.js" ./dist/wasm/go-wasm-exec.js