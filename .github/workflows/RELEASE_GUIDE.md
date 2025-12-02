# Release Management Guide

How to create releases with versioned Docker images.

## Creating a Release

### Method 1: Git Tag (Recommended)

```bash
git checkout main
git pull origin main
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

This triggers `release.yml` which builds all services and creates a GitHub Release.

### Method 2: Manual Dispatch

1. Actions → **Release - Build and Publish All Services**
2. Click **Run workflow**
3. Enter version: `v1.0.0` (must start with `v`)

## Version Tags

Tagging `v1.2.3` creates these image tags for each service:

- `v1.2.3` - Exact version
- `v1.2` - Minor version
- `v1` - Major version
- `latest` - Latest release

**Example:**

```
ghcr.io/{owner}/{repo}/portal-backend:v1.2.3
ghcr.io/{owner}/{repo}/portal-backend:v1.2
ghcr.io/{owner}/{repo}/portal-backend:v1
ghcr.io/{owner}/{repo}/portal-backend:latest
```

## Semantic Versioning

- **MAJOR** (v1.0.0 → v2.0.0): Breaking changes
- **MINOR** (v1.0.0 → v1.1.0): New features
- **PATCH** (v1.0.0 → v1.0.1): Bug fixes

## Using Release Images

**Production:**

```yaml
services:
  portal-backend:
    image: ghcr.io/{owner}/{repo}/portal-backend:v1.0.0
```

**Development:**

```yaml
services:
  portal-backend:
    image: ghcr.io/{owner}/{repo}/portal-backend:latest
```

## Rollback

Update docker-compose.yml to previous version:

```yaml
services:
  portal-backend:
    image: ghcr.io/{owner}/{repo}/portal-backend:v1.0.0
```

Then: `docker compose pull && docker compose up -d`

## Troubleshooting

**Workflow doesn't trigger:**

- Tag format must match `v*.*.*` pattern (e.g., `v1.0.0`, `v2.1.3`)
- Ensure tag was pushed: `git push origin v1.0.0`

**Security scan fails:**

- Review vulnerabilities in Security tab
- Update dependencies/base images
- Re-run workflow after fixes
