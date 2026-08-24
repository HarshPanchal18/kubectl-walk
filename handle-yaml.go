package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
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
		// The requested entry may legitimately be absent from
		// an individual resource when processing multiple resources.
		//
		// For example:
		//   pod.pod-1.metadata.labels
		//   pod.pod-2.metadata.labels: <not found>
		//
		// Do not fail the entire operation because one resource
		// does not contain the requested field.

		if errors.Is(err, ErrPathNotFound) {
			path = append(prefix, strings.Split(entry, ".")...)

			fmt.Fprintf(out, "%s: <not found>\n", strings.Join(path, "."))
			return nil
		}

		// Actual malformed/invalid entrypoint.
		return err
	}

	// Prefix identifies the resource when multiple resources are being processed
	path = append(prefix, path...)

	if tree {
		fmt.Fprintln(out, strings.Join(path, "."))
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
			name := part[:strings.Index(part, "[")]
			indexString := part[strings.Index(part, "[")+1 : strings.Index(part, "]")]

			index, err := strconv.Atoi(indexString)
			if err != nil {
				return nil, fmt.Errorf("invalid list index: %s", indexString)
			}

			// child object
			child := getMapValue(current, name)
			if child == nil {
				return nil, fmt.Errorf("%w: key %s", ErrPathNotFound, part)
			}

			// ensure list exists
			if child.Kind != yaml.SequenceNode || index < 0 || index >= len(child.Content) {
				return nil, fmt.Errorf("%w: index [%d] out of range for %s", ErrPathNotFound, index, name)
			}

			// move deeper into the list element
			current = child.Content[index]
			continue
		}

		// regular map key, no list
		next := getMapValue(current, part)
		if next == nil {
			return nil, fmt.Errorf("%w: key %s", ErrPathNotFound, part)
		}

		current = next
	}

	return current, nil
}

func watchResources(
	config *rest.Config,
	gvr schema.GroupVersionResource,
	namespaced bool,
	namespace, name, labelSelector string,
	out io.Writer,
) error {

	var previousOutput string
	firstRun := true

	for {
		nodes, err := loadYamlFromCluster(config, gvr, namespaced, namespace, name, labelSelector)

		if err != nil {
			return err
		}

		var current strings.Builder

		// Write output into buffer
		for _, node := range nodes {
			rendered, err := renderYaml(node, nil)
			if err != nil {
				return err
			}

			current.WriteString(rendered)
		}

		// Store the current buffer
		currentOutput := current.String()

		// Handle the separator between changes
		if firstRun {
			fmt.Fprint(out, currentOutput)
			firstRun = false
		} else if currentOutput != previousOutput { // Check buffer diff
			fmt.Fprintln(out, "---")
			fmt.Fprint(out, currentOutput)
		}

		// Store output for the next iteration
		previousOutput = currentOutput

		// Check every second
		time.Sleep(time.Second)
	}
}

func renderYaml(rootNode *yaml.Node, prefix []string) (string, error) {
	var buf bytes.Buffer

	if err := processYaml(rootNode, &buf, prefix); err != nil {
		return "", err
	}

	return buf.String(), nil
}
