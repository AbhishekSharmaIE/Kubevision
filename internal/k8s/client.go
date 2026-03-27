package k8s

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned"
)

// ClusterClient holds clients for a single cluster.
type ClusterClient struct {
	ID              string
	Name            string
	Environment     string
	Clientset       *kubernetes.Clientset
	MetricsClient   *metricsv1beta1.Clientset
	Config          *rest.Config
}

// ClusterManager manages connections to multiple Kubernetes clusters.
type ClusterManager struct {
	mu       sync.RWMutex
	clusters map[string]*ClusterClient
	pool     *pgxpool.Pool
}

// NewClusterManager creates an empty manager; clusters load from DB via LoadClusters.
func NewClusterManager(pool *pgxpool.Pool) *ClusterManager {
	return &ClusterManager{
		clusters: make(map[string]*ClusterClient),
		pool:     pool,
	}
}

// AddCluster registers a new cluster from kubeconfig bytes.
func (m *ClusterManager) AddCluster(ctx context.Context, id, name, env string, kubeconfigBytes []byte) error {
	config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfigBytes)
	if err != nil {
		return fmt.Errorf("invalid kubeconfig: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("creating clientset: %w", err)
	}
	metricsClient, err := metricsv1beta1.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("creating metrics client: %w", err)
	}
	if _, err := clientset.ServerVersion(); err != nil {
		return fmt.Errorf("cluster unreachable: %w", err)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clusters[id] = &ClusterClient{
		ID:            id,
		Name:          name,
		Environment:   env,
		Clientset:     clientset,
		MetricsClient: metricsClient,
		Config:        config,
	}
	return nil
}

// GetCluster returns a cluster client by ID.
func (m *ClusterManager) GetCluster(id string) (*ClusterClient, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.clusters[id]
	if !ok {
		return nil, fmt.Errorf("cluster %q not found", id)
	}
	return c, nil
}

// ListClusters returns registered in-memory clusters.
func (m *ClusterManager) ListClusters() []*ClusterClient {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*ClusterClient, 0, len(m.clusters))
	for _, c := range m.clusters {
		out = append(out, c)
	}
	return out
}

// LoadClusters reads cluster rows from Postgres and registers reachable clusters.
func (m *ClusterManager) LoadClusters(ctx context.Context) error {
	if m.pool == nil {
		return nil
	}
	rows, err := m.pool.Query(ctx, `
		SELECT id::text, name, environment, kubeconfig
		FROM clusters`)
	if err != nil {
		return fmt.Errorf("list clusters: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id, name, env string
		var kubeconfig []byte
		if err := rows.Scan(&id, &name, &env, &kubeconfig); err != nil {
			return fmt.Errorf("scan cluster: %w", err)
		}
		if err := m.AddCluster(ctx, id, name, env, kubeconfig); err != nil {
			slog.Warn("skip cluster registration", "id", id, "name", name, "error", err)
		}
	}
	return rows.Err()
}

// RemoveCluster drops a cluster from the in-memory map.
func (m *ClusterManager) RemoveCluster(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clusters, id)
}
