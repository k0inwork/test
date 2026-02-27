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
	fmt.Println("Starting PUM Unified Admin Distro...")

	// 1. Start the Bridge Agent in a goroutine
	go startBridgeAgent()

	// 2. Serve the embedded Apptron assets
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
	fmt.Println("Bridge Agent: Connecting to PUM Management Subnet...")
	// Bridge logic: Tunneling from Virtual Network to Local Subnet
}
