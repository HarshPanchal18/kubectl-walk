package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

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
			return nil, fmt.Errorf("error: %w: %s", ErrResourceNotFound, labelSelector)
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

func loadYamlFromFile(source string) ([]*yaml.Node, error) {
	var reader io.Reader

	switch {
	// cat Chart.yaml | kubectl walk -f -
	case source == "-":
		reader = os.Stdin

	// kubectl walk -f https://yaml-url
	case isURL(source):
		httpClient := &http.Client{Timeout: 10 * time.Second}
		response, err := httpClient.Get(source)
		if err != nil {
			return nil, fmt.Errorf("error: error fetching URL: %s %w", response.Request.URL, err)
		}

		if response.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("error: error fetching URL: status %s %w", response.Status, err)
		}
		defer response.Body.Close()

		reader = response.Body

	// kubectl walk -f file.yaml
	default:
		file, err := os.Open(source)
		if err != nil {
			return nil, fmt.Errorf("error: error reading file %s: %w", source, err)
		}

		defer file.Close()
		reader = file
	}

	decoder := yaml.NewDecoder(reader)

	var nodes []*yaml.Node

	for {
		var doc yaml.Node

		err := decoder.Decode(&doc)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if len(doc.Content) == 0 {
			continue
		}

		nodes = append(nodes, doc.Content[0])
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("error: no valid YAML documents found")
	}

	return nodes, nil
}

func getCurrentNamespace() string {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})

	ns, _, err := kubeConfig.Namespace()
	if err != nil || ns == "" {
		return "default"
	}

	return ns
}
