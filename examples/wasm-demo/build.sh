#!/bin/bash
# Build the WASM demo
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
EBITEN_DEMO="$SCRIPT_DIR/../ebiten-demo"

echo "Building WASM..."
cd "$EBITEN_DEMO"
GOOS=js GOARCH=wasm go build -o "$SCRIPT_DIR/main.wasm" .

echo "Copying wasm_exec.js..."
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" "$SCRIPT_DIR/"

echo "Done! Files in $SCRIPT_DIR:"
ls -lh "$SCRIPT_DIR"/*.wasm "$SCRIPT_DIR"/*.js "$SCRIPT_DIR"/*.html
