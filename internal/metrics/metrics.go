package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// EventOperations tracks the number of event operations
	EventOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_event_operations_total",
			Help: "The total number of event operations",
		},
		[]string{"operation", "status"},
	)

	// EventOperationDuration tracks the duration of event operations
	EventOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "app_event_operation_duration_seconds",
			Help:    "The duration of event operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// ActiveEventStreams tracks the number of active event streams
	ActiveEventStreams = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "app_active_event_streams",
			Help: "The number of currently active event streams",
		},
	)
)
