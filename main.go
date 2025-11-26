package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
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

func resolveKind(input string) string {
	switch strings.ToLower(input) {
	case "po", "pod", "pods":
		return "pod"
	case "deploy", "deployment", "deployments":
		return "deployment"
	case "svc", "service", "services":
		return "service"
	case "cm", "configmap", "configmaps":
		return "configmap"
	case "ing", "ingress", "ingresses":
		return "configmap"
	default:
		return input
	}
}

func parseResourceArg(arg string) (kind, name string, err error) {
	parts := strings.Split(arg, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid resource format, expected kind/name")
	}
	return resolveKind(parts[0]), parts[1], nil
}

func getResourceYaml(client *kubernetes.Clientset, ns, kind, name string) ([]byte, error) {
	ctx := context.Background()

	var obj runtime.Object
	var err error

	switch kind {
	case "deployment":
		obj, err = client.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})

	case "pod":
		obj, err = client.CoreV1().Pods(ns).Get(ctx, name, metav1.GetOptions{})

	case "pv":
		obj, err = client.CoreV1().PersistentVolumes().Get(ctx, name, metav1.GetOptions{})

	case "pvc":
		obj, err = client.CoreV1().PersistentVolumeClaims(ns).Get(ctx, name, metav1.GetOptions{})

	case "service":
		obj, err = client.CoreV1().Services(ns).Get(ctx, name, metav1.GetOptions{})

	case "configmap":
		obj, err = client.CoreV1().ConfigMaps(ns).Get(ctx, name, metav1.GetOptions{})

	case "ingress":
		obj, err = client.NetworkingV1().Ingresses(ns).Get(ctx, name, metav1.GetOptions{})

	default:
		return nil, fmt.Errorf("unsupported resource kind: %s", kind)
	}

	// Resource not found
	if err != nil {
		return nil, fmt.Errorf("'%s/%s' not found inside '%s' namespace", kind, name, ns)
	}

	// Detect GVK dynamically
	gvk := obj.GetObjectKind().GroupVersionKind()
	if gvk.Version == "" {
		kinds, _, _ := scheme.Scheme.ObjectKinds(obj)
		if len(kinds) > 0 { gvk = kinds[0] }
	}

	obj.GetObjectKind().SetGroupVersionKind(
		schema.GroupVersionKind{
			Group: gvk.Group,
			Version: gvk.Version,
			Kind: gvk.Kind,
		},
	)

	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)
	serializer := json.NewSerializerWithOptions(
		json.DefaultMetaFactory, scheme, scheme,
		json.SerializerOptions{ Yaml: true, Pretty: true, Strict: false },
	)
	return runtime.Encode(serializer, obj)
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

	if len(os.Args) < 2 {
        fmt.Println("Usage: walk <kind/name>")
        os.Exit(1)
    }

	kind, name, err := parseResourceArg(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ns := "default"

	yamlBytes, err := getResourceYaml(client, ns, kind, name)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Parse YAML into yaml.Node tree
	var yamlRoot yaml.Node
	yaml.Unmarshal(yamlBytes, &yamlRoot)

	walk(yamlRoot.Content[0], []string{})

}
