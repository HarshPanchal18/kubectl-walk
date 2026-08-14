package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

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

func isEmptyNode(node *yaml.Node) bool {
    switch node.Kind {
    case yaml.ScalarNode:
        return strings.TrimSpace(node.Value) == ""
    case yaml.MappingNode, yaml.SequenceNode:
        return len(node.Content) == 0
    default:
        return false
    }
}

func resolveEntryNode(root *yaml.Node, entry string) (*yaml.Node, []string, error) {
	if entry == "" {
		return root, []string{}, nil
	}

	node, err := findNodeByPath(root, entry)
	if err != nil {
		return nil, nil, err
	}

	return node, strings.Split(entry, "."), nil
}

func processYaml(rootNode *yaml.Node, out io.Writer, prefix []string) error {

	node, path, err := resolveEntryNode(rootNode, entry)
	if err != nil {
		return err
	}

	// Prefix identifies the resource when multiple resources are being processed
	path = append(prefix, path...)

	if tree {
		fmt.Println(strings.Join(path,"."))
		walkTree(node, "", out)
		return nil
	}

	walk(node, path, out, depth)
	return nil
}

func findNodeByPath(node *yaml.Node, entrypoint string) (*yaml.Node, error) {
	// Prevent accidentally passing of the document node
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		node = node.Content[0]
	}

	// get hierarchical segments
	parts := strings.Split(entrypoint, ".")
	current := node

	for _, part := range parts {

		// list index: containers[0]
		if strings.Contains(part, "[") {
			// extract name and the index between '[' and ']'
			name        := part[:strings.Index(part, "[")]
			indexString := part[strings.Index(part, "[") + 1:strings.Index(part, "]")]
			index, _    := strconv.Atoi(indexString)

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
            // return nil, fmt.Errorf("invalid format/entrypoint provided: %s", entrypoint)
            fmt.Println("invalid format/entrypoint provided:", entrypoint)
			os.Exit(0)
        }

		current = next
	}

	return current, nil
}