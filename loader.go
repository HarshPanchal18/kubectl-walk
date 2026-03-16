package main

import (
	"context"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

// FetchDynamic retrieves any Kubernetes resource using its kind, namespace, and name.
func FetchDynamicObject(
	ctx context.Context,
	restCfg *rest.Config,
	kind, ns, name string,
) (runtime.Object, error) {

	// Create a discovery client (needed for API group + version discovery)
	dc, err := discovery.NewDiscoveryClientForConfig(restCfg)
	if err != nil {
		return nil, fmt.Errorf("error creating discovery client: %w", err)
	}

	// RESTMapper caches API discovery and resolves Kind ↔︎ GVR
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	gvk, err := mapper.KindFor(schema.GroupVersionResource{Resource: kind})
	if err != nil {
		return nil, fmt.Errorf("error resolving GVK for %s: %w", kind, err)
	}

	// runtime-agnostic resource fetching
	dyn, err := dynamic.NewForConfig(restCfg)
	if err != nil {
		return nil, fmt.Errorf("error creating dynamic client: %w", err)
	}

	// identify resource
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil { return nil, err }

	var resource dynamic.ResourceInterface

	// Handle scopped object
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// namespaced resource
		resource = dyn.Resource(mapping.Resource).Namespace(ns)
	} else {
		// cluster-scoped resource
		resource = dyn.Resource(mapping.Resource)
	}

	// Fetch the object from Kubernetes
	obj, err := resource.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting %s/%s/%s (%s): %w", ns, kind, name, gvk.String(), err)
	}

	return obj, nil
}

func loadYamlFromCluster(kind, namespace, name string) (*yaml.Node, error) {

	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("error connecting Kubernetes: %w", err)
	}

	obj, err := FetchDynamicObject(context.TODO(), restConfig, kind, namespace, name)
	if err != nil {
		return nil, err
	}

	yamlBytes, err := serializeObject(obj)
	if err != nil {
		return nil, fmt.Errorf("serialization error: %w", err)
	}

	var yamlRoot yaml.Node
	if err := yaml.Unmarshal(yamlBytes, &yamlRoot); err != nil {
		return nil, err
	}

	return yamlRoot.Content[0], nil
}

func loadYamlFromFile(file string) (*yaml.Node, error) {
	yamlBytes, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", file, err)
	}

	var yamlRoot yaml.Node
	if err := yaml.Unmarshal(yamlBytes, &yamlRoot); err != nil {
		return nil, err
	}

	return yamlRoot.Content[0], nil
}