package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
)

//go:embed all:assets
var assets embed.FS

func main() {
	mode := os.Getenv("PUM_MODE")
	if mode == "" {
		mode = "mock"
	}

	fmt.Printf("Starting PUM Unified Admin Distro (Mode: %s)...\n", mode)

	// Custom Router
	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	// Mock API for projects (matches apptron.js urlFor logic)
	mux.HandleFunc("/projects/", func(w http.ResponseWriter, r *http.Request) {
		// e.g. /projects/test-project
		path := strings.TrimPrefix(r.URL.Path, "/projects/")

		if path == "" {
			if r.Method == "POST" {
				var req struct {
					Name string `json:"name"`
				}
				json.NewDecoder(r.Body).Decode(&req)
				projName := req.Name
				if projName == "" {
					projName = "New-Project"
				}

				w.Header().Set("Location", "/edit/"+projName)
				w.WriteHeader(http.StatusCreated)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"name":        projName,
					"description": "Mock newly created project",
					"updated_at":  "2024-03-21T10:00:00Z",
				})
				return
			}

			// GET /projects/
			projects := []map[string]interface{}{
				{
					"name":        "Main-Datacenter",
					"description": "Core infrastructure management",
					"updated_at":  "2024-03-20T10:00:00Z",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(projects)
			return
		}

		// Handle specific project requests like HEAD /projects/{projectName}
		if r.Method == "HEAD" {
			// Simulating that the project does not exist
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"name": path,
		})
	})

	mux.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var req struct {
				Name string `json:"name"`
			}
			json.NewDecoder(r.Body).Decode(&req)
			projName := req.Name
			if projName == "" {
				projName = "New-Project"
			}

			w.Header().Set("Location", "/edit/"+projName)
			w.WriteHeader(http.StatusCreated)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"name":        projName,
				"description": "Mock newly created project",
				"updated_at":  "2024-03-21T10:00:00Z",
			})
			return
		}
		projects := []map[string]interface{}{
			{
				"name":        "Main-Datacenter",
				"description": "Core infrastructure management",
				"updated_at":  "2024-03-20T10:00:00Z",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(projects)
	})

	// Mock Hanko/Auth
	mux.HandleFunc("/auth/user", func(w http.ResponseWriter, r *http.Request) {
		user := map[string]interface{}{
			"id": "1",
			"username": "admin",
			"email": "admin@example.com",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	})

	mux.HandleFunc("/auth/session", func(w http.ResponseWriter, r *http.Request) {
		session := map[string]interface{}{
			"is_valid": true,
			"claims": map[string]interface{}{
				"username": "admin",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(session)
	})

	// Assets handler with .html fallback and meta injection
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			http.Redirect(w, r, "/dashboard.html", http.StatusFound)
			return
		}

		contentStatic, _ := fs.Sub(assets, "assets")

		// Serve _env.html for editor and console directly in mock mode
		if strings.HasPrefix(path, "/edit/") || strings.HasPrefix(path, "/console/") {
			if data, err := fs.ReadFile(contentStatic, "_env.html"); err == nil {
				// Inject meta
				parts := strings.Split(path, "/")
				if len(parts) >= 3 {
					projName := parts[2]

					html := string(data)
					// Make sure it has a valid project meta for the mock apptron.js
					meta := fmt.Sprintf(`
    <meta name="auth-url" content="/auth">
    <meta name="project" content='{"name": "%s"}'>
    `, projName)
					html = strings.Replace(html, "<head>", "<head>"+meta, 1)

					w.Header().Set("Content-Type", "text/html; charset=utf-8")
					w.Write([]byte(html))
					return
				}
			}
		}

		// 1. Try exact path
		cleanPath := strings.TrimPrefix(path, "/")
		if data, err := fs.ReadFile(contentStatic, cleanPath); err == nil {
			contentType := "text/plain"
			if strings.HasSuffix(cleanPath, ".html") {
				contentType = "text/html; charset=utf-8"
				data = injectMeta(data)
			} else if strings.HasSuffix(cleanPath, ".js") {
				contentType = "application/javascript"
			} else if strings.HasSuffix(cleanPath, ".css") {
				contentType = "text/css"
			} else if strings.HasSuffix(cleanPath, ".svg") {
				contentType = "image/svg+xml"
			}
			w.Header().Set("Content-Type", contentType)
			w.Write(data)
			return
		}

		// 2. Try with .html suffix
		htmlPath := cleanPath + ".html"
		if data, err := fs.ReadFile(contentStatic, htmlPath); err == nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(injectMeta(data))
			return
		}

		http.NotFound(w, r)
	})

	port := os.Getenv("PUM_ADMIN_PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Admin Center available at http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

func injectMeta(data []byte) []byte {
	html := string(data)
	meta := `
    <meta name="auth-url" content="/auth">
    <meta name="project" content='{"name": "Local"}'>
    `
	return []byte(strings.Replace(html, "<head>", "<head>"+meta, 1))
}
