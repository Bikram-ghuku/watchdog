package config

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"watchdog.onebusaway.org/internal/app"
	"watchdog.onebusaway.org/internal/models"
	"watchdog.onebusaway.org/internal/server"
)

func TestLoadConfigFromFile(t *testing.T) {
	t.Run("ValidConfig", func(t *testing.T) {
		content := `[{
		"name": "Test Server", "id": 1,
		"oba_base_url": "https://test.example.com",
		"oba_api_key": "test-key",
		"gtfs_url": "https://gtfs.example.com",
		"trip_update_url": "https://trip.example.com",
		"vehicle_position_url": "https://vehicle.example.com",
		"gtfs_rt_api_key": "",
		"gtfs_rt_api_value": ""
		}]`
		tmpFile, err := os.CreateTemp("", "config-*.json")
		if err != nil {
			t.Fatalf("Failed to create temporary file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.Write([]byte(content)); err != nil {
			t.Fatalf("Failed to write to temporary file: %v", err)
		}
		tmpFile.Close()

		servers, err := LoadConfigFromFile(tmpFile.Name())
		if err != nil {
			t.Fatalf("loadConfigFromFile failed: %v", err)
		}

		if len(servers) != 1 {
			t.Fatalf("Expected 1 server, got %d", len(servers))
		}

		expected := models.ObaServer{
			Name:               "Test Server",
			ID:                 1,
			ObaBaseURL:         "https://test.example.com",
			ObaApiKey:          "test-key",
			GtfsUrl:            "https://gtfs.example.com",
			TripUpdateUrl:      "https://trip.example.com",
			VehiclePositionUrl: "https://vehicle.example.com",
			GtfsRtApiKey:       "",
			GtfsRtApiValue:     "",
		}

		if servers[0] != expected {
			t.Errorf("Expected server %+v, got %+v", expected, servers[0])
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		content := `{ this is not valid JSON }`
		tmpFile, err := os.CreateTemp("", "invalid-config-*.json")
		if err != nil {
			t.Fatalf("Failed to create temporary file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.Write([]byte(content)); err != nil {
			t.Fatalf("Failed to write to temporary file: %v", err)
		}
		tmpFile.Close()

		_, err = LoadConfigFromFile(tmpFile.Name())
		if err == nil {
			t.Errorf("Expected error with invalid JSON, got none")
		}
	})

	t.Run("NonExistentFile", func(t *testing.T) {
		_, err := LoadConfigFromFile("non-existent-file.json")
		if err == nil {
			t.Errorf("Expected error for non-existent file, got none")
		}
	})
}

func TestLoadConfigFromURL(t *testing.T) {
	t.Run("ValidResponse", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`[{"name": "Test Server",
			 "id": 1,
			 "oba_base_url": "https://test.example.com",
			 "oba_api_key": "test-key",
			 "gtfs_url": "https://gtfs.example.com",
			 "trip_update_url": "https://trip.example.com",
			 "vehicle_position_url": "https://vehicle.example.com",
			 "gtfs_rt_api_key": "",
			 "gtfs_rt_api_value": ""
			}]`))
		}))
		defer ts.Close()

		servers, err := LoadConfigFromURL(ts.URL, "user", "pass")
		if err != nil {
			t.Fatalf("loadConfigFromURL failed: %v", err)
		}

		if len(servers) != 1 {
			t.Fatalf("Expected 1 server, got %d", len(servers))
		}

		expected := models.ObaServer{
			Name:               "Test Server",
			ID:                 1,
			ObaBaseURL:         "https://test.example.com",
			ObaApiKey:          "test-key",
			GtfsUrl:            "https://gtfs.example.com",
			TripUpdateUrl:      "https://trip.example.com",
			VehiclePositionUrl: "https://vehicle.example.com",
		}

		if servers[0] != expected {
			t.Errorf("Expected server %+v, got %+v", expected, servers[0])
		}
	})

	t.Run("ErrorResponse", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		_, err := LoadConfigFromURL(ts.URL, "", "")
		if err == nil {
			t.Errorf("Expected error with 500 response, got none")
		}
	})

	t.Run("InvalidJSONResponse", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{ this is not valid JSON }`))
		}))
		defer ts.Close()

		_, err := LoadConfigFromURL(ts.URL, "", "")
		if err == nil {
			t.Errorf("Expected error for invalid JSON response, got none")
		}
	})
	t.Run("InvalidURL", func(t *testing.T) {
		_, err := LoadConfigFromURL("://invalid-url", "", "")
		if err == nil || !strings.Contains(err.Error(), "failed to create request") {
			t.Errorf("Expected request creation error, got: %v", err)
		}
	})
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		configFile  string
		configURL   string
		expectError bool
	}{
		{"Valid local config", "config.json", "", false},
		{"Valid remote config", "", "http://example.com/config.json", false},
		{"Both config file and URL", "config.json", "http://example.com/config.json", true},
		{"No config provided", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(tt.name, flag.ContinueOnError)
			os.Args = []string{"cmd", "--config-file=" + tt.configFile, "--config-url=" + tt.configURL}

			_, _, err := parseFlags()

			if (err != nil) != tt.expectError {
				t.Errorf("Expected error: %v, got: %v", tt.expectError, err)
			}
		})
	}
}

func parseFlags() (string, string, error) {
	var (
		configFile = flag.String("config-file", "", "Path to a local JSON configuration file")
		configURL  = flag.String("config-url", "", "URL to a remote JSON configuration file")
	)
	flag.Parse()

	// Check if both configFile and configURL are empty
	if *configFile == "" && *configURL == "" {
		return "", "", fmt.Errorf("no configuration provided. Use --config-file or --config-url")
	}

	// Check if more than one configuration option is provided
	if (*configFile != "" && *configURL != "") || (*configFile != "" && len(flag.Args()) > 0) || (*configURL != "" && len(flag.Args()) > 0) {
		return "", "", fmt.Errorf("only one of --config-file, --config-url, or raw config params can be specified")
	}

	return *configFile, *configURL, nil
}

func TestValidateConfigFlags(t *testing.T) {
	tests := []struct {
		name        string
		configFile  string
		configURL   string
		extraArgs   []string
		expectError bool
	}{
		{"No config", "", "", nil, false},
		{"Valid local config", "config.json", "", nil, false},
		{"Valid remote config", "", "http://example.com/config.json", nil, false},
		{"Both config file and URL", "config.json", "http://example.com/config.json", nil, true},
		{"Config file with extra args", "config.json", "", []string{"extraArg"}, true},
		{"Config URL with extra args", "", "http://example.com/config.json", []string{"extraArg"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(tt.name, flag.ContinueOnError)
			var output bytes.Buffer
			flag.CommandLine.SetOutput(&output)

			configFile := flag.String("config-file", "", "Path to config file")
			configURL := flag.String("config-url", "", "URL to config")

			args := []string{"cmd"}
			if tt.configFile != "" {
				args = append(args, "--config-file="+tt.configFile)
			}
			if tt.configURL != "" {
				args = append(args, "--config-url="+tt.configURL)
			}
			args = append(args, tt.extraArgs...)

			os.Args = args
			flag.CommandLine.Parse(args[1:])

			err := ValidateConfigFlags(configFile, configURL)

			if (err != nil) != tt.expectError {
				t.Errorf("Expected error: %v, got: %v", tt.expectError, err)
			}

			if err != nil && !strings.Contains(err.Error(), "only one of --config-file or --config-url") {
				t.Errorf("Unexpected error message: %v", err)
			}
		})
	}
}

func TestRefreshConfig(t *testing.T) {
	app := newTestApplication(t)

	testLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

	var serverHitCount int
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverHitCount++

		user, pass, hasAuth := r.BasicAuth()
		if hasAuth && (user != "testuser" || pass != "testpass") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `[
					{
							"id": 999,
							"name": "Refreshed Test Server",
							"url": "https://refreshed.example.com",
							"api_key": "refreshed-key",
							"gtfs_url": "https://refreshed.example.com/gtfs.zip"
					}
			]`)
	}))
	defer mockServer.Close()

	originalConfig := make([]models.ObaServer, len(app.Config.Servers))
	copy(originalConfig, app.Config.Servers)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go RefreshConfig(ctx, mockServer.URL, "testuser", "testpass", app, testLogger, 100*time.Millisecond)

	time.Sleep(200 * time.Millisecond)

	if serverHitCount == 0 {
		t.Fatal("Mock server was never called")
	}

	updatedServers := app.Config.GetServers()

	if len(updatedServers) == 0 {
		t.Fatal("No servers found in updated configuration")
	}

	var found bool
	for _, s := range updatedServers {
		if s.ID == 999 && s.Name == "Refreshed Test Server" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Config not updated with refreshed server data. Original: %+v, Updated: %+v", originalConfig, updatedServers)
	}
}

func newTestApplication(t *testing.T) *app.Application {
	t.Helper()

	obaServer := models.NewObaServer(
		"Test Server",
		1,
		"https://test.example.com",
		"test-key",
		"",
		"",
		"",
		"",
		"",
		"",
	)

	cfg := server.NewConfig(
		4000,
		"testing",
		[]models.ObaServer{*obaServer},
	)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	return &app.Application{
		Config: *cfg,
		Logger: logger,
	}
}
