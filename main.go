package main

import (
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func walk(node *yaml.Node, path string) {
	switch node.Kind {

	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i].Value
			value := node.Content[i+1]

			newPath := key
			if path != "" {
				newPath = path + "." + key
			}
			walk(value, newPath)
		}

	case yaml.SequenceNode:
		for i, item := range node.Content {
			newPath := fmt.Sprintf("%s[%d]", path, i)
			walk(item, newPath)
		}

	default: // scaler
		fmt.Printf("%s: %s\n", path, node.Value)
	}
}

func main() {
	filename := flag.Arg(0)

	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("error reading file: %v\n", err)
		os.Exit(1)
	}

	var root yaml.Node
	if err := yaml.Unmarshal(content, &root); err != nil {
		fmt.Printf("error parsing YAML: %v\n", err)
		os.Exit(1)
	}

	walk(root.Content[0], "")
}
