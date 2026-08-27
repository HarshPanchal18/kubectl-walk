package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/pflag"
)

// Supported CLI flags
var (
	namespace      string
	entry          string
	file           string
	outputFile     string
	kubeConfigPath string
	grep           string
	find           string
	labelSelector  string
	help           bool
	pure           bool
	includeEmpty   bool
	keysOnly       bool
	tree           bool
	valuesOnly     bool
	showVersion    bool
	noPrefixes     bool
	completion     bool
	watch          bool
	depth          int
	interval       time.Duration
)

func prepareCliFlags() {
	pflag.StringVarP(&namespace, "namespace", "n", "", "Namespace of resource")
	pflag.StringVarP(&entry, "entry", "e", "", "Entrypoint of an object")
	pflag.StringVarP(&file, "file", "f", "", "YAML file to read regardless of Kubernetes resource")
	pflag.StringVarP(&outputFile, "output", "o", "", "Write inside file instead of stdin")
	pflag.StringVar(&kubeConfigPath, "kubeconfig", os.Getenv("HOME")+"/.kube/config", "Cluster Kubeconfig file")
	pflag.StringVarP(&grep, "grep", "g", "", "Filter output paths by value substring")
	pflag.StringVar(&find, "find", "", "Search paths by field name")
	pflag.StringVarP(&labelSelector, "selector", "l", "", "Label selector (e.g. app=nginx)")

	pflag.BoolVarP(&help, "help", "h", false, "Print help")
	pflag.BoolVarP(&pure, "pure", "p", false, "Strip auto-generated fields")
	pflag.BoolVarP(&includeEmpty, "all", "A", false, "Include empty values")
	pflag.BoolVarP(&keysOnly, "keys", "k", false, "Include keys only. Ignore values.")
	pflag.BoolVarP(&tree, "tree", "t", false, "Render YAML structure as tree")
	pflag.BoolVar(&valuesOnly, "values", false, "Include values only")
	pflag.BoolVarP(&showVersion, "version", "v", false, "Print plugin version")
	pflag.BoolVar(&noPrefixes, "no-prefixes", false, "Disable resource prefixes when walking multiple objects")
	pflag.BoolVar(&completion, "completion", false, "Print Bash completion script")
	pflag.BoolVarP(&watch, "watch", "w", false, "Watch the resource for change")

	pflag.IntVarP(&depth, "depth", "d", -1, "Depth of walking on keys")
	pflag.DurationVarP(&interval, "interval", "i", time.Second, "Polling interval for --watch streams")

	pflag.Parse()
}

func printUsage() {
	fmt.Println("Interpret Kubernetes Resources like a human — not like a machine")

	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println(" $ kubectl walk ns default")
	fmt.Println(" $ kubectl walk pod nginx")
	fmt.Println(" $ kubectl walk pod nginx -e spec.containers")
	fmt.Println(" $ kubectl walk pod nginx -e spec.containers -d 2")
	fmt.Println(" $ kubectl walk pod nginx --keys")
	fmt.Println(" $ kubectl walk pod nginx --values")
	fmt.Println(" $ kubectl walk svc svc-nginx --find port")
	fmt.Println(" $ kubectl walk node -l beta.kubernetes.io/arch=amd64 -e metadata.labels")
	fmt.Println(" $ kubectl walk -f file.yaml")
	fmt.Println(" $ cat file.yaml | kubectl walk -f")
	fmt.Println(" $ kubectl walk -f https://file-at-remote.yaml")
	fmt.Println(" $ kubectl walk deploy nginx-deployment -e status --find Replica -w")
	fmt.Println(" $ kubectl walk deploy nginx-deployment -e status --find Replica -wi 10s")
	fmt.Println()

	fmt.Println("Enable bash completion:")
	fmt.Println("\t $ source <(kubectl walk --completion)")
	fmt.Println()

	fmt.Println("Usage:")
	pflag.PrintDefaults()
}

func sanitizeArgs() {

	if help {
		printUsage()
		os.Exit(0)
	}

	if showVersion {
		fmt.Printf("kubectl-walk %s\n", version) // go build -ldflags="-X main.version=v1.0.0 -s -w"
		os.Exit(0)
	}

	argsCount := len(pflag.Args())

	if completion {
		if argsCount > 0 {
			fmt.Println("error: --completion should not be followed by any positional argument")
		} else {
			printCompletion()
		}
		os.Exit(0)
	}

	if argsCount > 0 && file != "" {
		fmt.Println("error: no positional arguments are allowed other than flags when using the -f <file>")
		os.Exit(1)
	}

	if !watch && pflag.CommandLine.Changed("interval") {
		fmt.Println("error: -i/--interval is only applicable with -w/--watch")
		os.Exit(1)
	}

	if watch && interval < 100*time.Millisecond {
		fmt.Println("interval must be at least 100ms")
		os.Exit(1)
	}

	if valuesOnly && keysOnly {
		fmt.Println("error: --values and --keys cannot be used together")
		os.Exit(1)
	}

	if tree && (keysOnly || valuesOnly || grep != "" || find != "") {
		fmt.Println("Warning: --tree ignores --keys, --values, --grep, --find")
	}

	if namespace == "" {
		namespace = getCurrentNamespace()
	}
}
