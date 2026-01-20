# MicroUI-Go WASM Demo

Browser-based demo of microui-go using Ebiten's WebAssembly support.

## Quick Start

```bash
# From this directory
go run serve.go
```

Opens http://localhost:8080 in your browser.

## Manual Build

If you need to rebuild the WASM binary:

```bash
# From the ebiten-demo directory
cd ../ebiten-demo
GOOS=js GOARCH=wasm go build -o ../wasm-demo/main.wasm .

# Copy wasm_exec.js (if needed)
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" ../wasm-demo/
```

Or use the build script:

```bash
./build.sh
```

## Serving

WASM files must be served over HTTP (not file://). Options:

1. **Go server** (included): `go run serve.go`
2. **Python**: `python -m http.server 8080`
3. **Node**: `npx serve .`

## Files

- `index.html` - Host page
- `main.wasm` - Compiled demo (~12MB)
- `wasm_exec.js` - Go WASM runtime
- `serve.go` - Simple HTTP server
- `build.sh` - Build script

## Browser Support

Requires WebAssembly support (all modern browsers).
