package storage

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Collator metrics.
	collatorMergeError = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "vu_collator_merge_error_total",
		Help: "Count of errors merging revisions from collator",
	}, []string{"version"})
)
