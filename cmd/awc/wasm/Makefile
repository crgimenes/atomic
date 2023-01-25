@all:
	GOOS=js GOARCH=wasm go build -o main.wasm
	cp "$(GOROOT)/misc/wasm/wasm_exec.js" ../assets/wasm_exec.js
	mv main.wasm ../assets/main.wasm
