package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

func resolveKind(input string) string {
	switch input {
	case "po", "pod", "pods":
		return "pod"
	case "deploy", "deployment", "deployments":
		return "deployment"
	case "svc", "service", "services":
		return "service"
	case "cm", "configmap", "configmaps":
		return "configmap"
	case "pv", "persistentvolume", "persistentvolumes":
		return "persistentvolume"
	case "pvc", "persistentvolumeclaim", "persistentvolumeclaims":
		return "persistentvolumeclaim"
	case "ing", "ingress", "ingresses":
		return "configmap"
	default:
		return input
	}
}

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

	gvr, err := mapper.ResourceFor(schema.GroupVersionResource{Group: gvk.Group, Version: gvk.Version, Resource: gvk.Kind})
	if err != nil {
		return nil, fmt.Errorf("error resolving GVR for %s: %w", kind, err)
	}

	// runtime-agnostic resource fetching
	dyn, err := dynamic.NewForConfig(restCfg)
	if err != nil {
		return nil, fmt.Errorf("error creating dynamic client: %w", err)
	}

	// Fetch the object from Kubernetes
	obj, err := dyn.Resource(gvr).Namespace(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting %s/%s/%s (%s): %w", ns, kind, name, gvr.String(), err)
	}

	return obj, nil
}

func serializeObject(obj runtime.Object) ([]byte, error) {
	scheme := runtime.NewScheme()
	serializer := json.NewSerializerWithOptions(
		json.DefaultMetaFactory, scheme, scheme,
		json.SerializerOptions{Yaml: true},
	)
	return runtime.Encode(serializer, obj)
}

func walk(node *yaml.Node, path []string) {
	switch node.Kind {

	case yaml.MappingNode: // YAML object
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

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

	kubeConfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error connecting Kubernetes: %v\n", err)
		os.Exit(1)
	}

	// var help bool
	var namespace string

	// pflag.BoolVarP(&help, "help", "h", false, "Print help")
	pflag.StringVarP(&namespace, "namespace", "n", "default", "Namespace of kind")
	pflag.Parse()

	args := pflag.Args()
	kind := resolveKind(strings.ToLower(args[0]))
	resourceName := strings.ToLower(args[1])

	obj, err := FetchDynamicObject(context.TODO(), restConfig, kind, namespace, resourceName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	yamlBytes, err := serializeObject(obj)
	if err != nil {
		fmt.Println("serialization error: ", err)
		os.Exit(1)
	}

	// Parse YAML into yaml.Node tree
	var yamlRoot yaml.Node
	yaml.Unmarshal(yamlBytes, &yamlRoot)

	walk(yamlRoot.Content[0], []string{})

}
