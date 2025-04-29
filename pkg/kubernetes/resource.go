package kubernetes

import (
	"context"
	"fmt"

	authv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ResourcesList retrieves a list of resources for the given GroupVersionKind and namespace
// and returns them as a marshaled string
func (k *Kubernetes) ResourcesList(ctx context.Context, gvk *schema.GroupVersionKind, namespace string) (string, error) {
	resources, err := k.resourcesList(ctx, gvk, namespace)
	if err != nil {
		return "", fmt.Errorf("failed to list resources: %w", err)
	}

	marshaled, err := marshal(resources.Items)
	if err != nil {
		return "", fmt.Errorf("failed to marshal resources: %w", err)
	}
	return marshaled, nil
}

func (k *Kubernetes) resourcesList(ctx context.Context, gvk *schema.GroupVersionKind, namespace string) (*unstructured.UnstructuredList, error) {
	gvr, err := k.GetGroupVersionResource(gvk)
	if err != nil {
		return nil, err
	}
	isNamespaced, _ := k.checkResourceNamespaced(gvk)
	if isNamespaced && k.checkResourceAccess(ctx, gvr, namespace, "list") && namespace == "" {
		namespace = k.configuredNamespace()
	}
	return k.dynamicClient.Resource(*gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
}

// GetGroupVersionResource returns the GroupVersionResource for a given GroupVersionKind
func (k *Kubernetes) GetGroupVersionResource(gvk *schema.GroupVersionKind) (*schema.GroupVersionResource, error) {
	if gvk == nil {
		return nil, fmt.Errorf("GroupVersionKind cannot be nil")
	}

	mapping, err := k.deferredDiscoveryRESTMapper.RESTMapping(
		schema.GroupKind{
			Group: gvk.Group,
			Kind:  gvk.Kind,
		},
		gvk.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get REST mapping: %w", err)
	}

	return &mapping.Resource, nil
}

func (k *Kubernetes) checkResourceNamespaced(gvk *schema.GroupVersionKind) (bool, error) {
	// Get the API resource list for the given group version
	apiResourceList, err := k.discoveryClient.ServerResourcesForGroupVersion(gvk.GroupVersion().String())
	if err != nil {
		return false, err
	}

	// Find the matching API resource by kind and return its namespaced status
	for _, apiResource := range apiResourceList.APIResources {
		if apiResource.Kind == gvk.Kind {
			return apiResource.Namespaced, nil
		}
	}

	// Return false if no matching resource is found
	return false, nil
}

// CheckResourceAccess verifies if the current user has permission to perform
// the specified verb on a resource in the given namespace
func (k *Kubernetes) checkResourceAccess(ctx context.Context, gvr *schema.GroupVersionResource, namespace, verb string) bool {
	review := &authv1.SelfSubjectAccessReview{
		Spec: authv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authv1.ResourceAttributes{
				Namespace: namespace,
				Verb:      verb,
				Group:     gvr.Group,
				Version:   gvr.Version,
				Resource:  gvr.Resource,
			},
		},
	}
	response, err := k.clientSet.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, review, metav1.CreateOptions{})
	if err != nil {
		return false
	}
	return response.Status.Allowed
}
