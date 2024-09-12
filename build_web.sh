#!/bin/bash
set -ex

curl https://raw.githubusercontent.com/golang/go/go1.23.1/misc/wasm/wasm_exec.js > wasm_exec.js
GOOS=js GOARCH=wasm go build -o main.wasm main.go
python -m http.server 8000