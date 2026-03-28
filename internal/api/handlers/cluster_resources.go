package handlers

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/AbhishekSharmaIE/Kubevision/internal/api/httputil"
	"github.com/AbhishekSharmaIE/Kubevision/internal/k8s"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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

// ListServices handles GET /clusters/:id/namespaces/:namespace/services.
func (h *ClusterResourcesHandler) ListServices(c *gin.Context) {
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

	list, err := cl.Clientset.CoreV1().Services(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadGateway, "kubernetes list services failed: "+err.Error())
		return
	}
	out := make([]gin.H, 0, len(list.Items))
	for _, s := range list.Items {
		ports := make([]gin.H, 0, len(s.Spec.Ports))
		for _, p := range s.Spec.Ports {
			ports = append(ports, gin.H{
				"name":       p.Name,
				"port":       p.Port,
				"protocol":   string(p.Protocol),
				"targetPort": p.TargetPort.String(),
				"nodePort":   p.NodePort,
			})
		}
		lbHosts := make([]string, 0)
		lbIPs := make([]string, 0)
		for _, ing := range s.Status.LoadBalancer.Ingress {
			if ing.Hostname != "" {
				lbHosts = append(lbHosts, ing.Hostname)
			}
			if ing.IP != "" {
				lbIPs = append(lbIPs, ing.IP)
			}
		}
		out = append(out, gin.H{
			"name":              s.Name,
			"namespace":         s.Namespace,
			"type":              string(s.Spec.Type),
			"clusterIP":         s.Spec.ClusterIP,
			"externalName":      s.Spec.ExternalName,
			"selector":          s.Spec.Selector,
			"ports":             ports,
			"loadBalancerHosts": lbHosts,
			"loadBalancerIPs":   lbIPs,
			"creationTimestamp": s.CreationTimestamp.Time.UTC().Format(time.RFC3339),
		})
	}
	httputil.DataJSONWithMeta(c, http.StatusOK, out, gin.H{"total": len(out)})
}

func ingressBackendSummary(b networkingv1.IngressBackend) gin.H {
	if b.Service != nil {
		h := gin.H{"kind": "Service", "name": b.Service.Name}
		if b.Service.Port.Number != 0 {
			h["port"] = b.Service.Port.Number
		}
		if b.Service.Port.Name != "" {
			h["portName"] = b.Service.Port.Name
		}
		return h
	}
	if b.Resource != nil {
		return gin.H{
			"kind": "Resource", "apiGroup": b.Resource.APIGroup, "resourceKind": b.Resource.Kind, "name": b.Resource.Name,
		}
	}
	return gin.H{}
}

// ListIngresses handles GET /clusters/:id/namespaces/:namespace/ingresses.
func (h *ClusterResourcesHandler) ListIngresses(c *gin.Context) {
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

	list, err := cl.Clientset.NetworkingV1().Ingresses(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadGateway, "kubernetes list ingresses failed: "+err.Error())
		return
	}
	out := make([]gin.H, 0, len(list.Items))
	for _, ing := range list.Items {
		var className string
		if ing.Spec.IngressClassName != nil {
			className = *ing.Spec.IngressClassName
		}
		rules := make([]gin.H, 0, len(ing.Spec.Rules))
		hosts := make([]string, 0)
		for _, r := range ing.Spec.Rules {
			if r.Host != "" {
				hosts = append(hosts, r.Host)
			}
			paths := make([]gin.H, 0)
			if r.HTTP != nil {
				paths = make([]gin.H, 0, len(r.HTTP.Paths))
				for _, p := range r.HTTP.Paths {
					pt := ""
					if p.PathType != nil {
						pt = string(*p.PathType)
					}
					paths = append(paths, gin.H{
						"path": p.Path, "pathType": pt, "backend": ingressBackendSummary(p.Backend),
					})
				}
			}
			rules = append(rules, gin.H{"host": r.Host, "paths": paths})
		}
		tls := make([]gin.H, 0, len(ing.Spec.TLS))
		for _, t := range ing.Spec.TLS {
			tls = append(tls, gin.H{"hosts": t.Hosts, "secretName": t.SecretName})
		}
		lbHosts := make([]string, 0)
		lbIPs := make([]string, 0)
		for _, li := range ing.Status.LoadBalancer.Ingress {
			if li.Hostname != "" {
				lbHosts = append(lbHosts, li.Hostname)
			}
			if li.IP != "" {
				lbIPs = append(lbIPs, li.IP)
			}
		}
		out = append(out, gin.H{
			"name":              ing.Name,
			"namespace":         ing.Namespace,
			"ingressClassName":  className,
			"hosts":             hosts,
			"rules":             rules,
			"tls":               tls,
			"loadBalancerHosts": lbHosts,
			"loadBalancerIPs":   lbIPs,
			"labels":            ing.Labels,
			"creationTimestamp": ing.CreationTimestamp.Time.UTC().Format(time.RFC3339),
		})
	}
	httputil.DataJSONWithMeta(c, http.StatusOK, out, gin.H{"total": len(out)})
}

const maxLogBytes = 1 << 20 // 1 MiB

// GetPod handles GET /clusters/:id/namespaces/:namespace/pods/:pod.
func (h *ClusterResourcesHandler) GetPod(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	ns := strings.TrimSpace(c.Param("namespace"))
	podName := strings.TrimSpace(c.Param("pod"))
	if ns == "" || podName == "" {
		httputil.ErrorJSON(c, http.StatusBadRequest, "namespace and pod name required")
		return
	}
	cl, ok := h.client(c, id)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 45*time.Second)
	defer cancel()

	p, err := cl.Clientset.CoreV1().Pods(ns).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		httputil.ErrorJSON(c, http.StatusNotFound, "pod not found")
		return
	}

	httputil.DataJSON(c, http.StatusOK, podToJSON(p))
}

func podToJSON(p *corev1.Pod) gin.H {
	restarts := int32(0)
	containers := make([]gin.H, 0, len(p.Spec.Containers))
	for _, c := range p.Spec.Containers {
		containers = append(containers, gin.H{"name": c.Name, "image": c.Image})
	}
	for _, cs := range p.Status.ContainerStatuses {
		restarts += cs.RestartCount
	}
	phase := string(p.Status.Phase)
	if phase == "" {
		phase = "Unknown"
	}
	conds := make([]gin.H, 0, len(p.Status.Conditions))
	for _, co := range p.Status.Conditions {
		conds = append(conds, gin.H{
			"type": string(co.Type), "status": string(co.Status), "reason": co.Reason, "message": co.Message,
		})
	}
	return gin.H{
		"name":              p.Name,
		"namespace":         p.Namespace,
		"uid":               string(p.UID),
		"labels":            p.Labels,
		"nodeName":          p.Spec.NodeName,
		"phase":             phase,
		"restarts":          restarts,
		"containers":        containers,
		"conditions":        conds,
		"creationTimestamp": p.CreationTimestamp.Time.UTC().Format(time.RFC3339),
	}
}

// GetPodLogs handles GET /clusters/:id/namespaces/:namespace/pods/:pod/logs (snapshot; no follow).
func (h *ClusterResourcesHandler) GetPodLogs(c *gin.Context) {
	if strings.EqualFold(c.Query("follow"), "true") {
		httputil.ErrorJSON(c, http.StatusNotImplemented, "log streaming (follow=true) is not implemented; omit follow for a tail snapshot")
		return
	}
	id := strings.TrimSpace(c.Param("id"))
	ns := strings.TrimSpace(c.Param("namespace"))
	podName := strings.TrimSpace(c.Param("pod"))
	if ns == "" || podName == "" {
		httputil.ErrorJSON(c, http.StatusBadRequest, "namespace and pod name required")
		return
	}
	tail := int64(100)
	if v := c.Query("tailLines"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 && n <= 10000 {
			tail = n
		}
	}
	container := strings.TrimSpace(c.Query("container"))

	cl, ok := h.client(c, id)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	opts := &corev1.PodLogOptions{
		TailLines:  &tail,
		Timestamps: true,
	}
	if container != "" {
		opts.Container = container
	}
	req := cl.Clientset.CoreV1().Pods(ns).GetLogs(podName, opts)
	stream, err := req.Stream(ctx)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadGateway, "could not open log stream: "+err.Error())
		return
	}
	defer stream.Close()

	buf, err := io.ReadAll(io.LimitReader(stream, maxLogBytes))
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadGateway, "could not read logs: "+err.Error())
		return
	}
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("X-Content-Type-Options", "nosniff")
	c.String(http.StatusOK, string(buf))
}

const maxDeploymentReplicas = 10000

// GetDeployment handles GET /clusters/:id/namespaces/:namespace/deployments/:deployment.
func (h *ClusterResourcesHandler) GetDeployment(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	ns := strings.TrimSpace(c.Param("namespace"))
	name := strings.TrimSpace(c.Param("deployment"))
	if ns == "" || name == "" {
		httputil.ErrorJSON(c, http.StatusBadRequest, "namespace and deployment name required")
		return
	}
	cl, ok := h.client(c, id)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 45*time.Second)
	defer cancel()

	dep, err := cl.Clientset.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			httputil.ErrorJSON(c, http.StatusNotFound, "deployment not found")
			return
		}
		httputil.ErrorJSON(c, http.StatusBadGateway, "kubernetes get deployment failed: "+err.Error())
		return
	}
	httputil.DataJSON(c, http.StatusOK, deploymentToJSON(dep))
}

type scaleDeploymentBody struct {
	Replicas *int32 `json:"replicas" binding:"required"`
}

// ScaleDeployment handles PATCH /clusters/:id/namespaces/:namespace/deployments/:deployment/scale.
func (h *ClusterResourcesHandler) ScaleDeployment(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	ns := strings.TrimSpace(c.Param("namespace"))
	name := strings.TrimSpace(c.Param("deployment"))
	if ns == "" || name == "" {
		httputil.ErrorJSON(c, http.StatusBadRequest, "namespace and deployment name required")
		return
	}
	var body scaleDeploymentBody
	if err := c.ShouldBindJSON(&body); err != nil || body.Replicas == nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, "json body with replicas (integer) is required")
		return
	}
	if *body.Replicas < 0 || *body.Replicas > maxDeploymentReplicas {
		httputil.ErrorJSON(c, http.StatusBadRequest, "replicas must be between 0 and 10000")
		return
	}
	cl, ok := h.client(c, id)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 45*time.Second)
	defer cancel()

	scale, err := cl.Clientset.AppsV1().Deployments(ns).GetScale(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			httputil.ErrorJSON(c, http.StatusNotFound, "deployment not found")
			return
		}
		httputil.ErrorJSON(c, http.StatusBadGateway, "kubernetes get deployment scale failed: "+err.Error())
		return
	}
	scale.Spec.Replicas = *body.Replicas
	updated, err := cl.Clientset.AppsV1().Deployments(ns).UpdateScale(ctx, name, scale, metav1.UpdateOptions{})
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadGateway, "kubernetes scale deployment failed: "+err.Error())
		return
	}
	httputil.DataJSON(c, http.StatusOK, gin.H{
		"name":      updated.Name,
		"namespace": updated.Namespace,
		"replicas":  updated.Spec.Replicas,
	})
}

// RestartDeployment handles POST /clusters/:id/namespaces/:namespace/deployments/:deployment/restart (rolling restart via template annotation).
func (h *ClusterResourcesHandler) RestartDeployment(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	ns := strings.TrimSpace(c.Param("namespace"))
	name := strings.TrimSpace(c.Param("deployment"))
	if ns == "" || name == "" {
		httputil.ErrorJSON(c, http.StatusBadRequest, "namespace and deployment name required")
		return
	}
	cl, ok := h.client(c, id)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 45*time.Second)
	defer cancel()

	dep, err := cl.Clientset.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			httputil.ErrorJSON(c, http.StatusNotFound, "deployment not found")
			return
		}
		httputil.ErrorJSON(c, http.StatusBadGateway, "kubernetes get deployment failed: "+err.Error())
		return
	}
	if dep.Spec.Template.Annotations == nil {
		dep.Spec.Template.Annotations = make(map[string]string)
	}
	dep.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().UTC().Format(time.RFC3339)
	_, err = cl.Clientset.AppsV1().Deployments(ns).Update(ctx, dep, metav1.UpdateOptions{})
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadGateway, "kubernetes restart deployment failed: "+err.Error())
		return
	}
	httputil.DataJSON(c, http.StatusOK, gin.H{
		"name":        dep.Name,
		"namespace":   dep.Namespace,
		"restartedAt": dep.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"],
	})
}

// ListDeployments handles GET /clusters/:id/namespaces/:namespace/deployments.
func (h *ClusterResourcesHandler) ListDeployments(c *gin.Context) {
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

	list, err := cl.Clientset.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadGateway, "kubernetes list deployments failed: "+err.Error())
		return
	}
	out := make([]gin.H, 0, len(list.Items))
	for _, d := range list.Items {
		ready, desired := deploymentReplicas(d)
		out = append(out, gin.H{
			"name":              d.Name,
			"namespace":         d.Namespace,
			"replicas":          desired,
			"readyReplicas":     ready,
			"creationTimestamp": d.CreationTimestamp.Time.UTC().Format(time.RFC3339),
			"labels":            d.Labels,
			"images":            deploymentImages(d),
		})
	}
	httputil.DataJSONWithMeta(c, http.StatusOK, out, gin.H{"total": len(out)})
}

func deploymentReplicas(d appsv1.Deployment) (ready, desired int32) {
	desired = 1 // API default when spec.replicas is omitted
	if d.Spec.Replicas != nil {
		desired = *d.Spec.Replicas
	}
	ready = d.Status.ReadyReplicas
	return ready, desired
}

func deploymentImages(d appsv1.Deployment) []string {
	var imgs []string
	for _, c := range d.Spec.Template.Spec.Containers {
		imgs = append(imgs, c.Image)
	}
	return imgs
}

func deploymentToJSON(d *appsv1.Deployment) gin.H {
	ready, desired := deploymentReplicas(*d)
	strategy := string(appsv1.RollingUpdateDeploymentStrategyType)
	if d.Spec.Strategy.Type != "" {
		strategy = string(d.Spec.Strategy.Type)
	}
	ru := gin.H{}
	if d.Spec.Strategy.RollingUpdate != nil {
		if d.Spec.Strategy.RollingUpdate.MaxSurge != nil {
			ru["maxSurge"] = d.Spec.Strategy.RollingUpdate.MaxSurge.String()
		}
		if d.Spec.Strategy.RollingUpdate.MaxUnavailable != nil {
			ru["maxUnavailable"] = d.Spec.Strategy.RollingUpdate.MaxUnavailable.String()
		}
	}
	conds := make([]gin.H, 0, len(d.Status.Conditions))
	for _, co := range d.Status.Conditions {
		conds = append(conds, gin.H{
			"type":               string(co.Type),
			"status":             string(co.Status),
			"reason":             co.Reason,
			"message":            co.Message,
			"lastTransitionTime": formatTimePtr(co.LastTransitionTime),
		})
	}
	return gin.H{
		"name":              d.Name,
		"namespace":         d.Namespace,
		"uid":               string(d.UID),
		"labels":            d.Labels,
		"annotations":       d.Annotations,
		"replicas":          desired,
		"readyReplicas":     ready,
		"updatedReplicas":   d.Status.UpdatedReplicas,
		"availableReplicas": d.Status.AvailableReplicas,
		"paused":            d.Spec.Paused,
		"strategy":          strategy,
		"rollingUpdate":     ru,
		"selector":          d.Spec.Selector,
		"images":            deploymentImages(*d),
		"conditions":        conds,
		"creationTimestamp": d.CreationTimestamp.Time.UTC().Format(time.RFC3339),
	}
}

// ListEvents handles GET /clusters/:id/events.
func (h *ClusterResourcesHandler) ListEvents(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	cl, ok := h.client(c, id)
	if !ok {
		return
	}
	limit := int64(100)
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	nsFilter := strings.TrimSpace(c.Query("namespace"))
	typeFilter := strings.TrimSpace(c.Query("type"))

	ctx, cancel := context.WithTimeout(c.Request.Context(), 45*time.Second)
	defer cancel()

	lo := metav1.ListOptions{Limit: limit}
	if nsFilter != "" {
		lo.FieldSelector = "involvedObject.namespace=" + nsFilter
	}

	list, err := cl.Clientset.CoreV1().Events(metav1.NamespaceAll).List(ctx, lo)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadGateway, "kubernetes list events failed: "+err.Error())
		return
	}
	out := make([]gin.H, 0, len(list.Items))
	for _, e := range list.Items {
		if typeFilter != "" && string(e.Type) != typeFilter {
			continue
		}
		out = append(out, gin.H{
			"type":           e.Type,
			"reason":         e.Reason,
			"message":        e.Message,
			"namespace":      e.Namespace,
			"involvedObject": gin.H{"kind": e.InvolvedObject.Kind, "name": e.InvolvedObject.Name, "namespace": e.InvolvedObject.Namespace},
			"firstTimestamp": formatTimePtr(e.FirstTimestamp),
			"lastTimestamp":  formatTimePtr(e.LastTimestamp),
			"count":          e.Count,
		})
	}
	httputil.DataJSONWithMeta(c, http.StatusOK, out, gin.H{"total": len(out)})
}

func formatTimePtr(t metav1.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Time.UTC().Format(time.RFC3339)
}
