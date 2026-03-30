package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

func main() {

	prepareCliFlags()

	// Check --help and --version ahead of further execution
	sanitizeArgs()

	var err error
	out := os.Stdout

	// stdin support: cat file.yaml | kubectl walk
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Println("error: error reading stdin:", err)
			return
		}

		var yamlRoot yaml.Node
		if err := yaml.Unmarshal(data, &yamlRoot); err != nil {
			fmt.Println("error: error parsing yaml:", err)
			return
		}

		rootNode := yamlRoot.Content[0]

		if err := processYaml(rootNode, out); err != nil {
			fmt.Println("error:", err)
		}
		return
	}

	// Create a file if -o provided
	if outputFile != "" {
		out, err = os.Create(outputFile)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer out.Close()
	}

	var rootNode *yaml.Node

	// Read from .yaml file
	if file != "" {
		rootNode, err = loadYamlFromFile(file)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		args := pflag.Args()
		if len(args) < 2 {
			fmt.Println("error: kind and name required")
			return
		}
		kind := resolveKind(strings.ToLower(args[0]))
		name := strings.ToLower(args[1])

		rootNode, err = loadYamlFromCluster(kind, namespace, name)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	if err := processYaml(rootNode, out); err != nil {
		fmt.Println(err)
	}
}