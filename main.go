package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/api/meta"
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
    case "svc", "service", "services":
        return "service"
    case "cm", "configmap", "configmaps":
        return "configmap"
    case "secret", "secrets":
        return "secret"
    case "ns", "namespace", "namespaces":
        return "namespace"
    case "no", "node", "nodes":
        return "node"
    case "ev", "event", "events":
        return "event"
    case "sa", "serviceaccount", "serviceaccounts":
        return "serviceaccount"
    case "ep", "endpoints":
        return "endpoints"
    case "sc", "storageclass", "storageclasses":
        return "storageclass"
    case "pv", "persistentvolume", "persistentvolumes":
        return "persistentvolume"
    case "pvc", "persistentvolumeclaim", "persistentvolumeclaims":
        return "persistentvolumeclaim"

	case "deploy", "deployment", "deployments":
		return "deployment"
	case "sts", "statefulset", "statefulsets":
		return "statefulset"
    case "ds", "daemonset", "daemonsets":
        return "daemonset"
    case "rs", "replicaset", "replicasets":
        return "replicaset"

	case "job", "jobs":
		return "job"
	case "cj", "cronjob", "cronjobs":
		return "cronjob"

	case "ing", "ingress", "ingresses":
		return "ingress"
	case "netpol", "networkpolicy", "networkpolicies":
		return "networkpolicy"

	case "eplice", "endpointsslice", "endpointslices":
		return "endpointsslice"

    case "role", "roles":
        return "role"
    case "rb", "rolebinding", "rolebindings":
        return "rolebinding"
    case "cr", "clusterrole", "clusterroles":
        return "clusterrole"
    case "crb", "clusterrolebinding", "clusterrolebindings":
        return "clusterrolebinding"

    case "hpa", "horizontalpodautoscaler", "horizontalpodautoscalers":
        return "horizontalpodautoscaler"

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

	// runtime-agnostic resource fetching
	dyn, err := dynamic.NewForConfig(restCfg)
	if err != nil {
		return nil, fmt.Errorf("error creating dynamic client: %w", err)
	}

	// identify resource
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil { return nil, err }

	var resourceInterface dynamic.ResourceInterface

	// Handle scopped object
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// namespaced resource
		resourceInterface = dyn.Resource(mapping.Resource).Namespace(ns)
	} else {
		// cluster-scoped resource
		resourceInterface = dyn.Resource(mapping.Resource)
	}

	// Fetch the object from Kubernetes
	obj, err := resourceInterface.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting %s/%s/%s (%s): %w", ns, kind, name, gvk.String(), err)
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

func findNodeByPath(node *yaml.Node, entrypoint string) (*yaml.Node, error) {
	// get hierarchical segments
	parts := strings.Split(entrypoint, ".")
	current := node

	for _, part := range parts {

		// list index: containers[0]
		if strings.Contains(part, "[") {
			// extract name and the index between '[' and ']'
			name := part[:strings.Index(part, "[")]
			indexString := part[strings.Index(part, "[") + 1:strings.Index(part, "]")]
			index, _ := strconv.Atoi(indexString)

			// child object
			child := getMapValue(current, name)
			if child == nil {
				return nil, fmt.Errorf("key %s not found", name)
			}

			// ensure list exists
			if child.Kind != yaml.SequenceNode || index >= len(child.Content) {
				return nil, fmt.Errorf("index [%d] out of range for %s", index, name)
			}

			// move deeper into the list element
			current = child.Content[index]
			continue
		}

		// regular map key, no list
        next := getMapValue(current, part)
        if next == nil {
            return nil, fmt.Errorf("key %s not found", part)
        }

		current = next
	}

	return current, nil
}

// mapping node: get value for key
func getMapValue(node *yaml.Node, key string) *yaml.Node {
    if node.Kind != yaml.MappingNode {
        return nil
    }

	// Content[0] = key1, Content[1] = value1
	// Content[1] = key2, Content[1] = value2...
    for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			// Value for a given key
            return node.Content[i+1]
        }
    }

    return nil
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

func stdinHasData() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
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
	var entry string
	var file string
	// var pure string

	// pflag.BoolVarP(&help, "help", "h", false, "Print help")
	pflag.StringVarP(&namespace, "namespace", "n", "default", "Namespace of kind")
	pflag.StringVarP(&entry, "entry", "e", "", "Entrypoint of object")
	pflag.StringVarP(&file, "file", "f", "", "YAML file to read regardless of kubernetes resource")
	// pflag.StringVarP(&pure, "pure", "p", "", "Strip auto-generated fields")
	pflag.Parse()

	// Parse YAML into yaml.Node tree
	var yamlRoot yaml.Node

	// Read from .yaml file
	if file != "" {
		yamlBytes, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading file %s: %v\n", file, err)
			os.Exit(1)
		}

		yaml.Unmarshal(yamlBytes, &yamlRoot)
		rootNode := yamlRoot.Content[0]
		walk(rootNode, []string{})
		return
	}

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

	yaml.Unmarshal(yamlBytes, &yamlRoot)
	rootNode := yamlRoot.Content[0]

	if entry == "" {
		walk(rootNode, []string{})
		return
	}

	rootNode, err = findNodeByPath(rootNode, entry)
	if err != nil {
		panic(err)
	}

	walk(rootNode, strings.Split(entry, "."))
}
