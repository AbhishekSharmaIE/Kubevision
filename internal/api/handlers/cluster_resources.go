package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/AbhishekSharmaIE/Kubevision/internal/api/httputil"
	"github.com/AbhishekSharmaIE/Kubevision/internal/k8s"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterResourcesHandler lists Kubernetes objects via a registered ClusterManager client.
type ClusterResourcesHandler struct {
	mgr *k8s.ClusterManager
}

// NewClusterResourcesHandler constructs ClusterResourcesHandler.
func NewClusterResourcesHandler(mgr *k8s.ClusterManager) *ClusterResourcesHandler {
	return &ClusterResourcesHandler{mgr: mgr}
}

func (h *ClusterResourcesHandler) client(c *gin.Context, clusterID string) (*k8s.ClusterClient, bool) {
	if _, err := uuid.Parse(clusterID); err != nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, "invalid cluster id")
		return nil, false
	}
	cl, err := h.mgr.GetCluster(clusterID)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusNotFound, "cluster not connected; check registration and reachability")
		return nil, false
	}
	return cl, true
}

// ListNodes handles GET /clusters/:id/nodes.
func (h *ClusterResourcesHandler) ListNodes(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	cl, ok := h.client(c, id)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 45*time.Second)
	defer cancel()

	nl, err := cl.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadGateway, "kubernetes list nodes failed: "+err.Error())
		return
	}

	out := make([]gin.H, 0, len(nl.Items))
	for _, n := range nl.Items {
		ready := nodeReadyStatus(n)
		roles := nodeRoles(n.Labels)
		out = append(out, gin.H{
			"name":              n.Name,
			"ready":             ready,
			"roles":             roles,
			"kubeletVersion":    n.Status.NodeInfo.KubeletVersion,
			"internalIP":        nodeInternalIP(n),
			"creationTimestamp": n.CreationTimestamp.Time.UTC().Format(time.RFC3339),
		})
	}
	httputil.DataJSONWithMeta(c, http.StatusOK, out, gin.H{"total": len(out)})
}

func nodeInternalIP(n corev1.Node) string {
	for _, a := range n.Status.Addresses {
		if a.Type == corev1.NodeInternalIP {
			return a.Address
		}
	}
	return ""
}

func nodeReadyStatus(n corev1.Node) bool {
	for _, c := range n.Status.Conditions {
		if c.Type == corev1.NodeReady {
			return c.Status == corev1.ConditionTrue
		}
	}
	return false
}

func nodeRoles(labels map[string]string) []string {
	var r []string
	if labels["node-role.kubernetes.io/control-plane"] != "" || labels["node-role.kubernetes.io/master"] != "" {
		r = append(r, "control-plane")
	}
	if labels["node-role.kubernetes.io/worker"] != "" {
		r = append(r, "worker")
	}
	if len(r) == 0 {
		r = append(r, "node")
	}
	return r
}

// ListNamespaces handles GET /clusters/:id/namespaces.
func (h *ClusterResourcesHandler) ListNamespaces(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	cl, ok := h.client(c, id)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 45*time.Second)
	defer cancel()

	list, err := cl.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadGateway, "kubernetes list namespaces failed: "+err.Error())
		return
	}
	out := make([]gin.H, 0, len(list.Items))
	for _, ns := range list.Items {
		out = append(out, gin.H{
			"name":   ns.Name,
			"phase":  string(ns.Status.Phase),
			"labels": ns.Labels,
		})
	}
	httputil.DataJSONWithMeta(c, http.StatusOK, out, gin.H{"total": len(out)})
}

// ListPods handles GET /clusters/:id/namespaces/:namespace/pods.
func (h *ClusterResourcesHandler) ListPods(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	ns := strings.TrimSpace(c.Param("namespace"))
	if ns == "" {
		httputil.ErrorJSON(c, http.StatusBadRequest, "namespace required")
		return
	}
	cl, ok := h.client(c, id)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 45*time.Second)
	defer cancel()

	list, err := cl.Clientset.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadGateway, "kubernetes list pods failed: "+err.Error())
		return
	}
	out := make([]gin.H, 0, len(list.Items))
	for _, p := range list.Items {
		restarts := int32(0)
		for _, cs := range p.Status.ContainerStatuses {
			restarts += cs.RestartCount
		}
		phase := string(p.Status.Phase)
		if phase == "" {
			phase = "Unknown"
		}
		node := p.Spec.NodeName
		out = append(out, gin.H{
			"name":              p.Name,
			"namespace":         p.Namespace,
			"phase":             phase,
			"node":              node,
			"restarts":          restarts,
			"creationTimestamp": p.CreationTimestamp.Time.UTC().Format(time.RFC3339),
		})
	}
	httputil.DataJSONWithMeta(c, http.StatusOK, out, gin.H{"total": len(out)})
}
