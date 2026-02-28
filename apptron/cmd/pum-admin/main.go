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
	"sync"
	"time"
)

//go:embed all:assets
var assets embed.FS

type Project struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Visibility  string    `json:"visibility"`
	UpdatedAt   time.Time `json:"updated_at"`
}

var (
	projects   = make(map[string]Project)
	projectsMu sync.RWMutex
)

func init() {
	projects["Main-Datacenter"] = Project{
		Name:        "Main-Datacenter",
		Description: "Core infrastructure management",
		Visibility:  "private",
		UpdatedAt:   time.Now(),
	}
}

func main() {
	mode := os.Getenv("PUM_MODE")
	if mode == "" {
		mode = "mock"
	}

	fmt.Printf("Starting PUM Unified Admin Distro (Mode: %s)...\n", mode)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	mux.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		projectsMu.RLock()
		defer projectsMu.RUnlock()

		if r.Method == http.MethodGet {
			var list []Project
			for _, p := range projects {
				list = append(list, p)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(list)
			return
		}

		if r.Method == http.MethodPost {
			var p Project
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			p.UpdatedAt = time.Now()
			projectsMu.RUnlock()
			projectsMu.Lock()
			projects[p.Name] = p
			projectsMu.Unlock()
			projectsMu.RLock()

			w.Header().Set("Location", "/edit/"+p.Name)
			w.WriteHeader(http.StatusCreated)
			return
		}
	})

	mux.HandleFunc("/projects/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/projects/")
		name = strings.Split(name, "?")[0]
		name = strings.TrimSuffix(name, "/")

		projectsMu.RLock()
		p, ok := projects[name]
		projectsMu.RUnlock()

		if r.Method == http.MethodHead {
			if ok {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
			return
		}

		if r.Method == http.MethodGet {
			if !ok {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(p)
			return
		}

		if r.Method == http.MethodPut {
			var update struct {
				Description string `json:"description"`
				Visibility  string `json:"visibility"`
			}
			if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			projectsMu.Lock()
			if p, ok := projects[name]; ok {
				p.Description = update.Description
				p.Visibility = update.Visibility
				p.UpdatedAt = time.Now()
				projects[name] = p
			}
			projectsMu.Unlock()
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == http.MethodDelete {
			projectsMu.Lock()
			delete(projects, name)
			projectsMu.Unlock()
			w.WriteHeader(http.StatusOK)
			return
		}
	})

	mux.HandleFunc("/auth/user", func(w http.ResponseWriter, r *http.Request) {
		user := map[string]interface{}{
			"id":       "1",
			"username": "admin",
			"email":    "admin@example.com",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	})

	mux.HandleFunc("/auth/session", func(w http.ResponseWriter, r *http.Request) {
		session := map[string]interface{}{
			"is_valid": true,
			"claims": map[string]interface{}{
				"username": "admin",
				"sub":      "1",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(session)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			http.Redirect(w, r, "/dashboard.html", http.StatusFound)
			return
		}

		contentStatic, _ := fs.Sub(assets, "assets")

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
	if strings.Contains(html, "<head>") {
		return []byte(strings.Replace(html, "<head>", "<head>"+meta, 1))
	}
	return data
}
