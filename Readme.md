# kubectl-walk

Lightweight CLI tool to flatten and inspect nested fields of Kubernetes objects or YAML files. Handy and useful during documenting important field(s) instead of digging nested block.

Supports any YAML that ever exists, not only for Kubernetes resource(s).

## CLI flags

| Flag | Meaning |
|---|---|
| `--all`, `-A` | Include empty object value. |
| `--entry`, `-e`| Dotted entrypoint (e.g. `spec.template`) to start walking from. |
| `--depth`, `-d`| Depth of walking down the path |
| `--file`, `-f`| Read YAML from file instead of the cluster. |
| `--find`, | Search in keys by substring |
| `--grep`, `-g`| Search in values by substring |
| `--help`, `-h`| Show help message. |
| `--kubeconfig` | Path to kubeconfig (default `$HOME/.kube/config`). |
| `--keys` | Include keys only. Ignoring values. |
| `--namespace`, `-n`| Namespace (default `default`) of resource. |
| `--output`, `-o`| Write output to a file. |
| `--pure`, `-p`| Strip auto-generated fields when walking. |
| `--tree`, `-t` | Show fields in tree branch structure. |
| `--values` | Include values only. Takes priority when provided with `--keys`. |

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
# Build the binary from the repository root
go build -o kubectl-walk

# Put binary into the $PATH
chmod +x kubectl-walk
mv kubectl-walk ~/.local/bin

# Make sure ~/.local/bin is in your PATH.
echo $PATH

# Or add it via,
export PATH=$PATH:~/.local/bin

# Verify kubectl detects the plugin
kubectl plugin list
```

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
```

## Usefulness

- **Discovering fields easier**
- **Documenting the configuration of one or more fields**
- **Writing OPA policies or Kyverno rules**
- **Writing jsonpath**
