package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/pflag"
)

// Supported CLI flags
var (
	help bool
	namespace string
	entry string
	file string
	outputFile string
	kubeConfigPath string
	pure bool
	depth int
	includeEmpty bool
	grep string
	keysOnly bool
	tree bool
	find string
	valuesOnly bool
	showVersion bool
)

func prepareCliFlags() {
	pflag.StringVarP(&namespace, "namespace", "n", "default", "Namespace of resource")
	pflag.StringVarP(&entry, "entry", "e", "", "Entrypoint of an object")
	pflag.StringVarP(&file, "file", "f", "", "YAML file to read regardless of kubernetes resource")
	pflag.StringVarP(&outputFile, "output", "o", "", "Write inside file instead of stdin")
	pflag.StringVar(&kubeConfigPath, "kubeconfig", os.Getenv("HOME") + "/.kube/config", "Cluster Kubeconfig file")
	pflag.StringVarP(&grep, "grep", "g", "", "Filter output paths by value substring")
	pflag.StringVar(&find, "find", "", "Search paths by field name")

	pflag.BoolVarP(&help, "help", "h", false, "Print help")
	pflag.BoolVarP(&pure, "pure", "p", false, "Strip auto-generated fields")
	pflag.BoolVarP(&includeEmpty, "all", "A", false, "Include empty values")
	pflag.BoolVar(&keysOnly, "keys", false, "Include keys only")
	pflag.BoolVarP(&tree, "tree", "t", false, "Render YAML structure as tree")
	pflag.BoolVar(&valuesOnly, "values", false, "Include values only")
	pflag.BoolVarP(&showVersion, "version", "v", false, "Print plugin version")

	pflag.IntVarP(&depth, "depth", "d", -1, "Depth of walking on keys")

	pflag.Parse()
}

func printUsage() {
	fmt.Println("Explore Kubernetes YAML like a human — not like a machine")

	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println(" $ kubectl walk pod -n ns-web-nginx nginx")
	fmt.Println(" $ kubectl walk pod -n ns-web-nginx nginx -e spec.containers")
	fmt.Println(" $ kubectl walk pod -n ns-web-nginx nginx --keys")
	fmt.Println(" $ kubectl walk pod -n ns-web-nginx nginx --values")
	fmt.Println(" $ kubectl walk pod -n ns-web-nginx nginx --tree")
	fmt.Println(" $ kubectl walk pod -n ns-web-nginx nginx --find image")
	fmt.Println(" $ kubectl walk pod -n ns-web-nginx nginx --grep nginx")
	fmt.Println(" $ kubectl walk -f file.yaml")
	fmt.Println()

	fmt.Println("Usage:")
	pflag.PrintDefaults()
}

func sanitizeArgs() {
	if valuesOnly && keysOnly {
		fmt.Println("error: --values and --keys cannot be used together")
		os.Exit(1)
	}

	if tree && (keysOnly || valuesOnly || grep != "" || find != "") {
		fmt.Println("Warning: --tree ignores --keys, --values, --grep, --find")
	}

	if help {
		printUsage()
		return
	}

	if showVersion {
		fmt.Printf("kubectl-walk %s (built with Go %s)\n", version, runtime.Version()) // go build -ldflags="-X main.version=v1.0.0 -s -w"
		return
	}
}