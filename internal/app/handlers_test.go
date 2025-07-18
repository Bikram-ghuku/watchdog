package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"watchdog.onebusaway.org/internal/models"
	"watchdog.onebusaway.org/internal/server"
)

func TestHealthcheckHandler(t *testing.T) {
	t.Run("returns 200 OK when servers are configured", func(t *testing.T) {
		app := &Application{
			Config: server.Config{
				Env:     "testing",
				Servers: []models.ObaServer{{ID: 1, Name: "Test Server"}},
			},
			Version: "test-version",
		}

		rr := httptest.NewRecorder()
		request, err := http.NewRequest(http.MethodGet, "/v1/healthcheck", nil)
		if err != nil {
			t.Fatal(err)
		}

		app.healthcheckHandler(rr, request)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		var resp struct {
			Status      string `json:"status"`
			Environment string `json:"environment"`
			Version     string `json:"version"`
			Servers     int    `json:"servers"`
			Ready       bool   `json:"ready"`
		}

		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Status != "available" {
			t.Errorf("expected status 'available', got %q", resp.Status)
		}
		if resp.Environment != "testing" {
			t.Errorf("expected environment 'testing', got %q", resp.Environment)
		}
		if resp.Version != "test-version" {
			t.Errorf("expected version 'test-version', got %q", resp.Version)
		}
		if resp.Servers != 1 {
			t.Errorf("expected servers 1, got %d", resp.Servers)
		}
		if !resp.Ready {
			t.Errorf("expected ready true, got false")
		}
	})

	t.Run("returns 500 when no servers configured", func(t *testing.T) {
		app := &Application{
			Config: server.Config{
				Env:     "testing",
				Servers: []models.ObaServer{},
			},
			Version: "test-version",
		}

		rr := httptest.NewRecorder()
		request, err := http.NewRequest(http.MethodGet, "/v1/healthcheck", nil)
		if err != nil {
			t.Fatal(err)
		}

		app.healthcheckHandler(rr, request)

		if status := rr.Code; status != http.StatusInternalServerError {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusInternalServerError)
		}

		var resp struct {
			Status      string `json:"status"`
			Environment string `json:"environment"`
			Version     string `json:"version"`
			Servers     int    `json:"servers"`
			Ready       bool   `json:"ready"`
		}

		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Ready {
			t.Errorf("expected ready false, got true")
		}
		if resp.Servers != 0 {
			t.Errorf("expected servers 0, got %d", resp.Servers)
		}
		if resp.Status != "available" {
			t.Errorf("expected status 'available', got %q", resp.Status)
		}
		if resp.Environment != "testing" {
			t.Errorf("expected environment 'testing', got %q", resp.Environment)
		}
		if resp.Version != "test-version" {
			t.Errorf("expected version 'test-version', got %q", resp.Version)
		}
	})
}
