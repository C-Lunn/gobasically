#!/usr/bin/fish

GOARCH=wasm GOOS=js go build -o index.wasm wasm_main.go
