# Helm Chart Repository

This directory contains the Helm chart repository for kube-mcp.

## Usage

To use this repository:

```bash
helm repo add kube-mcp https://wrkode.github.io/kube-mcp/charts
helm repo update
helm install kube-mcp kube-mcp/kube-mcp
```

## Repository Structure

- `index.yaml` - Chart repository index (auto-generated)
- `kube-mcp-*.tgz` - Packaged Helm charts

## Publishing

Charts are automatically published to this repository via GitHub Actions when:
- Charts are updated in the `charts/` directory
- A new release is published
- The workflow is manually triggered

## Manual Publishing

To manually publish a chart:

```bash
helm package ../kube-mcp
helm repo index . --url https://wrkode.github.io/kube-mcp/charts
```
