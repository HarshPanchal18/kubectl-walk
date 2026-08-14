package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	prepareCliFlags()

	// Check --help and --version ahead of further execution
	sanitizeArgs()

	var err error
	out := os.Stdout

	// Create a file if -o provided
	if outputFile != "" {
		out, err = os.Create(outputFile)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer out.Close()
	}

	var nodes []*yaml.Node

	// Read from .yaml file
	if file != "" {
		nodes, err = loadYamlFromFile(file)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		args := pflag.Args()

		if len(args) < 2 && labelSelector == "" {
			fmt.Println(fmt.Errorf("error: insufficient parameter provided"))
			return
		}

		if labelSelector != "" && len(args) >= 2 {
			fmt.Println("error: name cannot be provided when a selector is specified")
			return
		}

		kind := strings.ToLower(args[0])
		var name string
		if labelSelector == "" {
			name = strings.ToLower(args[1])
		}

		restConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			fmt.Println("error: error while connecting Kubernetes:", err)
			return
		}

		gvr, namespaced, err := resolveGVR(restConfig, kind)
		if err != nil {
			fmt.Println("error:", err)
			return
		}

		nodes, err = loadYamlFromCluster(restConfig, gvr, namespaced, namespace, name, labelSelector)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	// Expand the Kubernets List object before applying kind.name prefix
	nodes = expandListNodes(nodes)

	// Apply prefix if there are multiple resources
	usePrefixes := len(nodes) > 1 && !noPrefixes

	// YAML is collected, make it flat
	for _, node := range nodes {
		var prefix []string

		if usePrefixes {
			prefix = resourcePrefix(node)
		}

		if err := processYaml(node, out, prefix); err != nil {
			fmt.Println(err)
		}

		fmt.Fprintln(out, "---")
	}
}
