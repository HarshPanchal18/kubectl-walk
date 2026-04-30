package main

import (
	"context"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func loadYamlFromCluster(config *rest.Config, gvr schema.GroupVersionResource, namespaced bool, namespace, name, labelSelector string) ([]*yaml.Node, error) {

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	var nodes []*yaml.Node
	var items []unstructured.Unstructured

	var resource dynamic.ResourceInterface

	if namespaced { // Ignore -n silently
		resource = dynamicClient.Resource(gvr).Namespace(namespace)
	} else {
		resource = dynamicClient.Resource(gvr)
	}

	if labelSelector != "" {
		list, err := resource.List(context.TODO(), metav1.ListOptions{LabelSelector: labelSelector})
		if err != nil {
			return nil, err
		}

		if len(list.Items) == 0 {
			return nil, fmt.Errorf("error: no resources found for selector: %s", labelSelector)
		}

		items = list.Items
		nodes = make([]*yaml.Node, len(items))
	} else {
		object, err := resource.Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("error: %w", err)
		}

		items = []unstructured.Unstructured{*object}
		nodes = make([]*yaml.Node, len(items))
	}

	for i, obj := range items {
		yamlBytes, err := yaml.Marshal(obj.Object)
		if err != nil {
			return nil, fmt.Errorf("serialization error: %w", err)
		}

		var yamlRoot yaml.Node
		if err := yaml.Unmarshal(yamlBytes, &yamlRoot); err != nil {
			return nil, err
		}

		if len(yamlRoot.Content) == 0 {
			continue
		}

		nodes[i] = yamlRoot.Content[0]
	}

	return nodes, nil
}

func loadYamlFromFile(file string) ([]*yaml.Node, error) {
	yamlBytes, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", file, err)
	}

	var yamlRoot yaml.Node
	if err := yaml.Unmarshal(yamlBytes, &yamlRoot); err != nil {
		return nil, err
	}

	return []*yaml.Node{yamlRoot.Content[0]}, nil
}

func getCurrentNamespace() string {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeConfig   := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})

	ns, _, err := kubeConfig.Namespace()
	if err != nil || ns == "" {
		return "default"
	}

	return ns
}