# kubectl-walk - explore Kubernetes YAML like a human — not like a machine

Lightweight CLI tool to flatten, filters, and visualizes nested fields of Kubernetes objects or YAML files. Handy and useful during documenting important field(s) instead of digging nested block.

Supports any YAML that ever exists, not only for Kubernetes resource(s).

## Why this exists?

### 1. Explore Unknown Resources (CRDs)

When working with tools like Argo CD, Istio, or Prometheus, resources can be deeply nested and undocumented.

```bash
kubectl walk applications my-app --tree
```

- Quickly understand structure without digging docs.

---

### 2. Discover Field Paths (Better than guessing JSONPath)

Instead of trial-and-error:

```bash
kubectl get pod nginx -o jsonpath='{.spec.containers[0].image}'
```

Do:

```bash
kubectl walk pod nginx --find image
```

Instantly discover the correct path:

```yaml
spec.containers[0].image
```

---

### 3. Debug Kubernetes Resources Faster

Find what’s wrong without scrolling massive YAML:

```bash
kubectl walk pod nginx --grep CrashLoop
```

- Extract only relevant lines.

---

### 4. Focus on Specific Sections

Avoid noise:

```bash
kubectl walk deployment api -e spec.template.spec
```

- Zoom directly into container configs.

---

### 5. Flatten YAML for Readability

Kubernetes YAML is deeply nested. Flatten it:

```bash
kubectl walk pod nginx
```

Output:

```yaml
...
spec.containers[0].image: nginx
spec.containers[0].name: web
...
```

- Much easier to scan than raw YAML.

---

### 6. Extract Only Keys (Schema Discovery)

```bash
kubectl walk pod nginx --keys
```

- Useful for:
  - understanding structure
  - building automation
  - writing policies

---

### 7. Extract Only Values (Quick Inspection)

```bash
kubectl walk pod nginx --values
```

- Great for:
  - quick audits
  - checking runtime values
  - piping into scripts

---

### 8. Visual Tree View

```bash
kubectl walk pod nginx --tree
```

Understand hierarchy instantly:

```bash
spec
└── containers
    └── [0]
        ├── name
        └── image
```

---

### 9. Search by Field Name

```bash
kubectl walk deployment api --find replicas
```

- Locate specific configuration fields quickly.

---

### 10. Remove Kubernetes Noise

Kubernetes adds a lot of auto-generated fields.

```bash
kubectl walk pod nginx --pure
```

- Removes:
  - `status`
  - `resourceVersion`
  - `managedFields`
  - etc.

---

### 11. Work with Local Files

```bash
kubectl walk -f deployment.yaml
```

- No cluster required.

---

### 12. Pipe from kubectl (Unix-style)

```bash
kubectl get pod nginx -o yaml | kubectl walk --find image
```

- Seamless CLI workflows.

---

### 13. Control Depth of Exploration

```bash
kubectl walk pod nginx --depth 2
```

- Avoid overwhelming output.

---

### 14. Debug Generated YAML (Helm, Kustomize)

Works great with tools like:

- Helm
- Kustomize

```bash
helm template . | kubectl walk --tree
```

- Inspect generated manifests easily.

---

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
