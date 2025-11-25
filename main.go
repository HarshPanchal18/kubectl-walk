package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// getClientset - loads kubeconfig and returns a Kubernetes clientset
func getClientset() (*kubernetes.Clientset, error) {
    kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
    config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
    if err != nil {
        return nil, err
    }
    return kubernetes.NewForConfig(config)
}

func walk(node *yaml.Node, path []string) {
	switch node.Kind {

	case yaml.MappingNode: // YAML object
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i + 1]

			walk(valueNode, append(path, keyNode.Value))
		}

	case yaml.SequenceNode: // YAML list: arr[0], arr[1], ...
		for i, item := range node.Content {
			walk(item, append(path, fmt.Sprintf("\b[%d]", i)))
		}

	default: // reached a scaler value (tail)
		fmt.Printf("%s: %s\n", strings.Join(path, "."), node.Value)
	}
}

func main() {

	client, err := getClientset()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Kubernetes client: %v\n", err)
		os.Exit(1)
	}

	resource, err := client.AppsV1().Deployments("default").Get(
		context.TODO(),
		"console",
		metav1.GetOptions{},
	)

	if err != nil {
		panic(err)
	}

	// Convert Kubernetes object to YAML
	scheme := runtime.NewScheme()
	serializer := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme, scheme)
	yamlBytes, err := runtime.Encode(serializer, resource)
	if err != nil {
		panic(err)
	}

	// Parse YAML into yaml.Node tree
	var yamlRoot yaml.Node
	if err := yaml.Unmarshal(yamlBytes, &yamlRoot); err != nil {
		panic(err)
	}

	walk(yamlRoot.Content[0], []string{})
}
