# Helm Charts Repository Setup

This document describes how to set up GitHub Pages for publishing Helm charts from the `gh-pages` branch.

## Overview

kube-mcp uses a `gh-pages` branch in the same repository to host Helm charts via GitHub Pages. This follows Helm best practices and keeps chart artifacts separate from source code while maintaining everything in one repository.

## Repository Setup Instructions

### Step 1: Create the gh-pages Branch

If the `gh-pages` branch doesn't exist yet, create it:

```bash
# Clone your repository
git clone https://github.com/YOUR_USERNAME/kube-mcp.git
cd kube-mcp

# Create and checkout gh-pages branch
git checkout --orphan gh-pages
git rm -rf .

# Create an initial commit (empty is fine)
git commit --allow-empty -m "Initial commit for charts repository"
git push origin gh-pages
```

### Step 2: Configure GitHub Pages

1. Go to your repository on GitHub
2. Navigate to **Settings** → **Pages**
3. Configure:
   - **Source**: Deploy from a branch
   - **Branch**: `gh-pages` → `/ (root)`
4. Click **Save**

Your charts will be available at: `https://YOUR_USERNAME.github.io/kube-mcp/`

**Note**: It may take a few minutes for GitHub Pages to be available after first setup.

### Step 4: Verify Setup

1. Create a test release or manually trigger the charts workflow
2. Check that the workflow completes successfully
3. Verify the charts repository has:
   - `index.yaml` file
   - Chart `.tgz` files
4. Test adding the repository:
   ```bash
   helm repo add kube-mcp https://YOUR_USERNAME.github.io/kube-mcp-charts
   helm repo update
   helm search repo kube-mcp
   ```

## Workflow Behavior

The charts workflow (`charts.yaml`) automatically:

1. **Triggers on**:
   - Release published (stable or pre-release)
   - Manual workflow dispatch

2. **Process**:
   - Checks out the source repository
   - Updates chart versions
   - Packages the Helm chart
   - Checks out the charts repository (`gh-pages` branch)
   - Generates/updates `index.yaml`
   - Commits and pushes charts to the charts repository

3. **Repository URL**:
   - Default: `https://YOUR_USERNAME.github.io/kube-mcp-charts`
   - Configurable via `CHARTS_REPO` secret

## Using the Charts Repository

Once set up, users can add and use the charts repository:

```bash
# Add the repository
helm repo add kube-mcp https://YOUR_USERNAME.github.io/kube-mcp-charts

# Update repository cache
helm repo update

# Search for charts
helm search repo kube-mcp

# Install kube-mcp
helm install kube-mcp kube-mcp/kube-mcp

# Install specific version
helm install kube-mcp kube-mcp/kube-mcp --version 1.0.0
```

## Troubleshooting

### Workflow Fails: "Repository not found"

- Ensure the charts repository exists and is public
- Verify the `CHARTS_REPO` secret matches the repository name
- Check that the repository has a `gh-pages` branch

### Workflow Fails: "Permission denied"

- Ensure GitHub Actions has write permissions to the charts repository
- The workflow uses `GITHUB_TOKEN` which has access to repositories in the same organization/user account

### Charts Not Appearing on GitHub Pages

- Verify GitHub Pages is enabled for the `gh-pages` branch
- Check that `index.yaml` exists in the repository root
- Wait a few minutes for GitHub Pages to rebuild

### Wrong Repository URL

- Update the `CHARTS_REPO` secret in your main repository
- The URL format is: `https://OWNER.github.io/REPO_NAME`

## Migration from docs/charts

If you're migrating from the old `docs/charts` location:

1. The old charts in `docs/charts/` can be removed (they're no longer used)
2. Existing users will need to update their repository URL:
   ```bash
   helm repo remove kube-mcp  # Remove old repo
   helm repo add kube-mcp https://YOUR_USERNAME.github.io/kube-mcp-charts
   helm repo update
   ```

## Repository Structure

The charts repository (`gh-pages` branch) will contain:

```
kube-mcp-charts/
├── index.yaml          # Chart repository index (auto-generated)
├── kube-mcp-1.0.0.tgz # Chart packages
├── kube-mcp-1.0.1.tgz
└── ...
```

**Important**: Do not manually edit files in the charts repository. The workflow automatically manages all files.
