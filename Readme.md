# `kubectl-walk` - interpret Kubernetes resources like a human — not like a machine

A `kubectl krew` plugin to flatten, filters, and visualizes nested fields of Kubernetes objects or YAML files. Handy and useful during documenting important field(s) instead of digging nested block.

Supports any YAML that ever exists, not only for Kubernetes resource(s).

## Installation

```bash
kubectl krew install walk
```

## Bash Completion

`kubectl-walk` provides Bash completion for **resource types, resource names,
namespaces, flags, and file paths.**

- Generate the completion script with:

  ```bash
  kubectl walk --completion
  ```

- Enable for the current shell

  ```bash
  source <(kubectl walk --completion)
  ```

- Enable permanently

  Create the Bash completion file:

  ```bash
  mkdir -p ~/.local/share/bash-completion/completions

  kubectl walk --completion \
    > ~/.local/share/bash-completion/completions/kubectl-walk
  ```

  Then restart your shell or reload Bash completion.

## CLI flags

| Flag | Meaning |
|---|---|
|  -A, --all                 | Include empty values |
|      --completion          | Print Bash completion script |
|  -d, --depth int           | Depth of walking on keys (default -1) |
|  -e, --entry string        | Entrypoint of an object |
|  -f, --file string         | YAML file to read regardless of Kubernetes resource |
|      --find string         | Search paths by field name |
|  -g, --grep string         | Filter output paths by value substring |
|  -h, --help                | Print help |
|  -k, --keys                | Include keys only. Ignore values. |
|      --kubeconfig string   | Cluster Kubeconfig file (default "/home/harsh/.kube/config") |
|  -n, --namespace string    | Namespace of resource (defaults to `default`) |
|      --no-prefixes         | Disable resource prefixes when walking multiple objects |
|  -o, --output string       | Write inside file instead of stdin |
|  -p, --pure                | Strip auto-generated fields |
|  -l, --selector string     | Label selector for Kubernetes resources (e.g. app=nginx) |
|  -t, --tree                | Render YAML structure as tree |
|      --values              | Include values only. Takes priority when provided with `--keys` |
|  -v, --version             | Print plugin version |

## Usage

```bash
# Inspect a live Kubernetes object (requires kubeconfig)
kubectl walk pod nginx -n default

# Flatten only a subtree of containers
kubectl walk pod nginx -n default -e spec.containers

# Read from a YAML file instead of the cluster
kubectl walk -f example.yaml -e spec.containers
# Or, via
cat example.yaml | kubectl walk -e spec.containers

# Write output to a file
kubectl walk -f example.yaml -e spec.containers -o output.txt

# Inspect cluster-scoped resources (i.e. namespace)
kubectl walk ns default
```

## Overview

### Standard YAML

```yaml
apiVersion: v1
kind: Pod
metadata:
    name: nginx-pod
    namespace: default
spec:
    containers:
      - image: nginx
        imagePullPolicy: Always
        name: nginx-pod
```

### Transformation

```yaml
apiVersion: v1
kind: Pod
metadata.name: nginx-pod
metadata.namespace: default
spec.containers[0].image: nginx
spec.containers[0].imagePullPolicy: Always
spec.containers[0].name: nginx-pod
```

## Build and Test on local

```bash
# Build the binary from the repository root and move to the ~/.local/bin directory
make apply

# Make sure ~/.local/bin is in your PATH.
echo $PATH

# Or add it via,
export PATH=$PATH:~/.local/bin

# Verify kubectl detects the plugin
kubectl plugin list
```
