package kflap

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// ResourceInfo holds information about a Kubernetes resource
type ResourceInfo struct {
	Name            string
	Type            string
	Namespace       string
	ResourceVersion int64
	Changes         int64
}

// Monitor handles polling Kubernetes resources
type Monitor struct {
	config        Config
	dynamicClient dynamic.Interface
	clientset     *kubernetes.Clientset
	resources     map[string]*ResourceInfo // key: namespace/type/name
	mu            sync.RWMutex
}

// NewMonitor creates a new resource monitor
func NewMonitor(config Config) (*Monitor, error) {
	// Load kubeconfig
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	restConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading kubeconfig: %v", err)
	}

	// Create dynamic client for generic resource access
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating dynamic client: %v", err)
	}

	// Create clientset for discovery
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating clientset: %v", err)
	}

	return &Monitor{
		config:        config,
		dynamicClient: dynamicClient,
		clientset:     clientset,
		resources:     make(map[string]*ResourceInfo),
	}, nil
}

// Poll fetches current resource versions and calculates deltas
func (m *Monitor) Poll() error {
	ctx := context.Background()

	// Discover API resources
	apiResources, err := m.discoverResources()
	if err != nil {
		return fmt.Errorf("error discovering resources: %v", err)
	}

	// Determine which namespaces to query
	namespaces := m.config.Namespaces
	if len(namespaces) == 0 {
		// Get all namespaces
		nsList, err := m.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("error listing namespaces: %v", err)
		}
		for _, ns := range nsList.Items {
			namespaces = append(namespaces, ns.Name)
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Poll each resource type
	for _, apiResource := range apiResources {
		gvr := schema.GroupVersionResource{
			Group:    apiResource.Group,
			Version:  apiResource.Version,
			Resource: apiResource.Name,
		}

		// Handle namespaced vs cluster-scoped resources
		if apiResource.Namespaced {
			for _, ns := range namespaces {
				if err := m.pollNamespacedResource(ctx, gvr, ns, apiResource.Kind); err != nil {
					// Log error but continue with other resources
					continue
				}
			}
		} else {
			if err := m.pollClusterResource(ctx, gvr, apiResource.Kind); err != nil {
				// Log error but continue with other resources
				continue
			}
		}
	}

	return nil
}

func (m *Monitor) pollNamespacedResource(ctx context.Context, gvr schema.GroupVersionResource, namespace, kind string) error {
	list, err := m.dynamicClient.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		m.updateResourceInfo(item.GetName(), kind, namespace, item.GetResourceVersion())
	}

	return nil
}

func (m *Monitor) pollClusterResource(ctx context.Context, gvr schema.GroupVersionResource, kind string) error {
	list, err := m.dynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		m.updateResourceInfo(item.GetName(), kind, "", item.GetResourceVersion())
	}

	return nil
}

func (m *Monitor) updateResourceInfo(name, resourceType, namespace, versionStr string) {
	key := fmt.Sprintf("%s/%s/%s", namespace, resourceType, name)

	version, err := strconv.ParseInt(versionStr, 10, 64)
	if err != nil {
		return
	}

	if existing, ok := m.resources[key]; ok {
		changes := existing.Changes
		if version != existing.ResourceVersion {
			changes++
		}
		m.resources[key] = &ResourceInfo{
			Name:            name,
			Type:            resourceType,
			Namespace:       namespace,
			ResourceVersion: version,
			Changes:         changes,
		}
	} else {
		// First time seeing this resource
		m.resources[key] = &ResourceInfo{
			Name:            name,
			Type:            resourceType,
			Namespace:       namespace,
			ResourceVersion: version,
		}
	}
}

// GetResources returns a copy of current resource information
func (m *Monitor) GetResources() []*ResourceInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*ResourceInfo, 0, len(m.resources))
	for _, info := range m.resources {
		// Create a copy
		infoCopy := *info
		result = append(result, &infoCopy)
	}
	return result
}

type apiResourceInfo struct {
	Group      string
	Version    string
	Name       string
	Kind       string
	Namespaced bool
}

func (m *Monitor) discoverResources() ([]apiResourceInfo, error) {
	var result []apiResourceInfo

	// ServerPreferredResources returns only the preferred (latest stable) version
	// per resource kind, avoiding deprecation warnings from older versions.
	apiResourceLists, err := m.clientset.Discovery().ServerPreferredResources()
	if err != nil {
		// Partial errors are common with CRDs, continue with what we have
	}

	// Filter by configured resource types if specified
	resourceFilter := make(map[string]bool)
	if len(m.config.Resources) > 0 {
		for _, r := range m.config.Resources {
			resourceFilter[r] = true
		}
	}

	// Resources whose replacement is a different resource name; skip to avoid
	// deprecation warnings (e.g. v1 endpoints → discovery.k8s.io/v1 endpointslices).
	deprecatedResources := map[string]map[string]bool{
		"v1": {"endpoints": true, "componentstatuses": true},
	}

	for _, apiResourceList := range apiResourceLists {
		gv, err := schema.ParseGroupVersion(apiResourceList.GroupVersion)
		if err != nil {
			continue
		}

		for _, apiResource := range apiResourceList.APIResources {
			// Skip subresources
			if len(apiResource.Name) > 0 && (apiResource.Name[len(apiResource.Name)-1] == '/' ||
				contains(apiResource.Name, "/")) {
				continue
			}

			// Skip deprecated resources superseded by newer API versions
			if deprecated, ok := deprecatedResources[apiResourceList.GroupVersion]; ok {
				if deprecated[apiResource.Name] {
					continue
				}
			}

			// Check if resource has "list" verb (indicates it's listable)
			if !hasVerb(apiResource.Verbs, "list") {
				continue
			}

			// Apply resource filter if configured
			if len(resourceFilter) > 0 {
				if !resourceFilter[apiResource.Name] && !resourceFilter[apiResource.Kind] {
					continue
				}
			}

			result = append(result, apiResourceInfo{
				Group:      gv.Group,
				Version:    gv.Version,
				Name:       apiResource.Name,
				Kind:       apiResource.Kind,
				Namespaced: apiResource.Namespaced,
			})
		}
	}

	return result, nil
}

func hasVerb(verbs metav1.Verbs, verb string) bool {
	for _, v := range verbs {
		if v == verb {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
