package engine

import (
	"sync"

	"github.com/volcanion-company/volcanion-stress-test-tool/internal/metrics"
)

var (
	sharedTestCollector     *metrics.Collector
	sharedTestCollectorOnce sync.Once
)

// getSharedTestCollector returns a singleton collector shared across all test files
// This prevents duplicate Prometheus metric registration errors
func getSharedTestCollector() *metrics.Collector {
	sharedTestCollectorOnce.Do(func() {
		sharedTestCollector = metrics.NewCollector()
	})
	return sharedTestCollector
}
