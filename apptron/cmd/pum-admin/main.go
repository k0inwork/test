package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
)

//go:embed all:assets
var assets embed.FS

func main() {
	mode := os.Getenv("PUM_MODE")
	if mode == "" {
		mode = "mock"
	}

	fmt.Printf("Starting PUM Unified Admin Distro (Mode: %s)...\n", mode)

	// The Bridge Agent always runs, but its behavior changes based on mode
	go startBridgeAgent(mode)

	// Health endpoint for testing
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	// Serve the embedded Apptron assets
	contentStatic, _ := fs.Sub(assets, "assets")
	http.Handle("/", http.FileServer(http.FS(contentStatic)))

	port := os.Getenv("PUM_ADMIN_PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Admin Center available at http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func startBridgeAgent(mode string) {
	if mode == "mock" {
		fmt.Println("Bridge Agent: Running in MOCK mode (Internal loopback only).")
		// In mock mode, the bridge might just serve a static response for the WASM ping test
		http.HandleFunc("/bridge-test", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "Bridge MOCK Response")
		})
	} else {
		fmt.Println("Bridge Agent: Connecting to LIVE PUM Management Subnet...")
		// Real L3 bridging logic
	}
}
