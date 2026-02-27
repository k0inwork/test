package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
)

// In a real build, we would embed the result of scripts/build_distro.sh
// For this prototype, we embed the source assets directory.
//go:embed all:assets
var assets embed.FS

func main() {
	mode := os.Getenv("PUM_MODE")
	if mode == "" {
		mode = "mock"
	}

	fmt.Printf("Starting PUM Unified Admin Distro (Mode: %s)...\n", mode)

	if mode == "live" {
		go startBridgeAgent()
	} else {
		fmt.Println("Running in MOCK mode. Bridge Agent disabled.")
	}

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

func startBridgeAgent() {
	fmt.Println("Bridge Agent: Connecting to LIVE PUM Management Subnet...")
	// Bridge logic: Tunneling from Virtual Network to Local Subnet
}
