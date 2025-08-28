//go:build integration

package integration

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"testing"

	"watchdog.onebusaway.org/internal/geo"
	"watchdog.onebusaway.org/internal/gtfs"
)

// TestDownloadGTFSBundles verifies that GTFS bundles can be downloaded successfully
// for all configured servers. It runs a subtest for each server in parallel,
// and checks that the downloaded file is created without error.
func TestDownloadGTFSBundles(t *testing.T) {
	if len(integrationServers) == 0 {
		t.Skip("No servers found in config")
	}

	staticStore := gtfs.NewStaticStore()
	realtimeStore := gtfs.NewRealtimeStore()
	boundingBoxStore := geo.NewBoundingBoxStore()
	logger := slog.Default()
	client := &http.Client{}
	gtfsService := gtfs.NewGtfsService(staticStore,realtimeStore,boundingBoxStore,logger,client)
	ctx := context.Background()
	for _, server := range integrationServers {
		srv := server
		t.Run(fmt.Sprintf("ServerID_%d", srv.ID), func(t *testing.T) {
			t.Parallel()
			staticBundle,err := gtfsService.DownloadGTFSBundle(ctx,srv.GtfsUrl, srv.ID,20)
			if err != nil {
				t.Errorf("failed to download GTFS bundle for server %d : %v", srv.ID, err)
				return
			}
			err = gtfsService.StoreGTFSBundle(staticBundle,server.ID)
			if err != nil {
				t.Errorf("failed to store GTFS bundle for server %d : %v", srv.ID, err)
				return
			}
			t.Logf("GTFS bundle downloaded successfully for server %d", srv.ID)
		})
	}
}
