package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// ObaApiStatus API Status (up/down)
	ObaApiStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "oba_api_status",
			Help: "Status of the OneBusAway API Server (0 = not working, 1 = working)",
		},
		[]string{"server_id", "server_url"},
	)
)

var (
	BundleEarliestExpirationGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gtfs_bundle_days_until_earliest_expiration",
		Help: "Number of days until the earliest GTFS bundle expiration",
	}, []string{"agency_id"})

	BundleLatestExpirationGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gtfs_bundle_days_until_latest_expiration",
		Help: "Number of days until the latest GTFS bundle expiration",
	}, []string{"agency_id"})
)

var (
	AgenciesInStaticGtfs = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "oba_agencies_in_static_gtfs",
		Help: "Number of agencies in the static GTFS file",
	})

	AgenciesInCoverageEndpoint = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "oba_agencies_in_coverage_endpoint",
		Help: "Number of agencies in the agencies-with-coverage endpoint",
	})

	AgenciesDifference = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "oba_agencies_difference",
		Help: "Difference between the number of agencies in the static GTFS file and the agencies-with-coverage endpoint",
	})
)
