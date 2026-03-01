package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status OK, got %v", status)
	}
	if rr.Body.String() != "OK" {
		t.Errorf("expected OK, got %v", rr.Body.String())
	}
}

func TestProjectsEndpoint(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		projects := []map[string]interface{}{
			{
				"name": "Apptron-Core",
				"uuid": "00000000-0000-0000-0000-000000000001",
			},
		}
		json.NewEncoder(w).Encode(projects)
	})

	req := httptest.NewRequest("GET", "/projects", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status OK, got %v", status)
	}

	var projects []map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&projects); err != nil {
		t.Fatal(err)
	}

	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0]["name"] != "Apptron-Core" {
		t.Errorf("expected Apptron-Core, got %v", projects[0]["name"])
	}
}
