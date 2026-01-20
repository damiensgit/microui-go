// Simple HTTP server for testing the WASM demo locally.
// Run: go run serve.go
// Then open: http://localhost:8080
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

func main() {
	port := "8080"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	// Serve current directory
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)

	url := fmt.Sprintf("http://localhost:%s", port)
	fmt.Printf("Serving WASM demo at %s\n", url)
	fmt.Println("Press Ctrl+C to stop")

	// Try to open browser
	go func() {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			cmd = exec.Command("cmd", "/c", "start", url)
		case "darwin":
			cmd = exec.Command("open", url)
		default:
			cmd = exec.Command("xdg-open", url)
		}
		cmd.Run()
	}()

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
