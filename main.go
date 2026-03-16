package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

func main() {

	prepareCliFlags()

	if help {
		printUsage()
		return
	}

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
			fmt.Println("kind and name required")
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