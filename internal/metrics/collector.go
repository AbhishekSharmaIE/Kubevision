package metrics

import (
	"context"
	"log/slog"
	"time"

	"github.com/AbhishekSharmaIE/Kubevision/internal/k8s"
	"github.com/redis/go-redis/v9"
)

// Collector pulls and caches cluster metrics (stub loop until Prometheus wiring lands).
type Collector struct {
	clusters *k8s.ClusterManager
	redis    *redis.Client
}

// NewCollector builds a metrics collector.
func NewCollector(clusters *k8s.ClusterManager, redisClient *redis.Client) *Collector {
	return &Collector{clusters: clusters, redis: redisClient}
}

// Start runs a lightweight background ticker (placeholder for Prometheus queries).
func (c *Collector) Start(ctx context.Context) {
	t := time.NewTicker(30 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			n := len(c.clusters.ListClusters())
			slog.Debug("metrics collector tick", "registered_clusters", n)
		}
	}
}
