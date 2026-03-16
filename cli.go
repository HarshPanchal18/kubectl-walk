package main

import (
	"fmt"
	"os"

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
)

func prepareCliFlags() {
	pflag.BoolVarP(&help, "help", "h", false, "Print help")
	pflag.StringVarP(&namespace, "namespace", "n", "default", "Namespace of resource")
	pflag.StringVarP(&entry, "entry", "e", "", "Entrypoint of an object")
	pflag.StringVarP(&file, "file", "f", "", "YAML file to read regardless of kubernetes resource")
	pflag.StringVarP(&outputFile, "output", "o", "", "Write inside file instead of stdin")
	pflag.StringVarP(&kubeConfigPath, "kubeconfig", "c", os.Getenv("HOME") + "/.kube/config", "Cluster Kubeconfig file")
	pflag.BoolVarP(&pure, "pure", "p", false, "Strip auto-generated fields")
	pflag.IntVarP(&depth, "depth", "d", -1, "Depth of walking")
	pflag.BoolVarP(&includeEmpty, "--all", "A", false, "Include empty values")
	pflag.Parse()
}

func printUsage() {
	fmt.Println("Flatten nested objects of the YAML.")
	fmt.Println("$ kubectl walk pod nginx --entry spec.containers")
	fmt.Print(
		"spec.containers[0].image: nginx\n" +
		"spec.containers[0].imagePullPolicy: Always\n" +
		"spec.containers[0].name: nginx-pod\n" +
		"spec.containers[0].terminationMessagePath: /dev/termination-log\n" +
		"spec.containers[0].terminationMessagePolicy: File\n" +
		"spec.containers[0].volumeMounts[0].mountPath: /var/run/secrets/kubernetes.io/serviceaccount\n" +
		"spec.containers[0].volumeMounts[0].name: kube-api-access-vvbkx\n" +
		"spec.containers[0].volumeMounts[0].readOnly: true\n")

	fmt.Println("Usage:")
	pflag.PrintDefaults()
}