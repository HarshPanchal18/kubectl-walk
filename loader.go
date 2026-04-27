package main

import (
	"context"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

func loadYamlFromCluster(config *rest.Config, gvr schema.GroupVersionResource, namespaced bool, namespace, name string) (*yaml.Node, error) {

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	var resource dynamic.ResourceInterface

	if namespaced { // Ignore -n silently
		resource = dynamicClient.Resource(gvr).Namespace(namespace)
	} else {
		resource = dynamicClient.Resource(gvr)
	}

	object, err := resource.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}

	yamlBytes, err := yaml.Marshal(object.Object)
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