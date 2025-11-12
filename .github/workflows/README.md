# GitHub Actions Docker Image Build Workflows

Automatically builds and publishes Docker images to GitHub Container Registry (ghcr.io) when code is merged to main.

## Available Workflows

| Workflow | Service | Image |
|----------|---------|-------|
| `build-api-server-go.yml` | API Server | `ghcr.io/{owner}/{repo}/api-server-go` |
| `build-audit-service.yml` | Audit Service | `ghcr.io/{owner}/{repo}/audit-service` |
| `build-policy-decision-point.yml` | Policy Decision Point | `ghcr.io/{owner}/{repo}/policy-decision-point` |
| `build-consent-engine.yml` | Consent Engine | `ghcr.io/{owner}/{repo}/consent-engine` |
| `build-orchestration-engine.yml` | Orchestration Engine | `ghcr.io/{owner}/{repo}/orchestration-engine` |
| `release.yml` | All Services | Builds all services with version tags |

## How It Works

**Triggers:**
- Push to main (when service code changes)
- Manual dispatch from GitHub Actions UI

**Process:**
1. Builds Docker image
2. Tags with `latest` and commit SHA
3. Scans for vulnerabilities (Trivy)
4. Publishes to GHCR

**Image Tags:**
- `latest` - Latest build from main
- `{branch}-{sha}` - Specific commit (e.g., `main-abc123`)

## Quick Test

```bash
# Test local build
cd exchange
docker build -f consent-engine/Dockerfile \
  --build-arg SERVICE_PATH=consent-engine \
  --build-arg BUILD_VERSION=test \
  --build-arg BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  --build-arg GIT_COMMIT=test \
  -t consent-engine:test .

# Test image runs
docker run --rm -p 8081:8081 \
  -e ENVIRONMENT=local -e PORT=8081 \
  consent-engine:test
```

## Security Scanning

All workflows include Trivy scanning:
- Scans images after build
- Fails on CRITICAL/HIGH vulnerabilities
- Results in GitHub Security tab

**View results:** Repository → Security → Code scanning alerts

## Using Published Images

Update `docker-compose.yml`:

```yaml
services:
  consent-engine:
    image: ghcr.io/{owner}/{repo}/consent-engine:latest
    # Remove 'build:' section
```

Then:
```bash
docker compose pull
docker compose up -d
```

## Troubleshooting

**Workflow doesn't trigger:**
- Check service directory files changed
- Verify branch is `main`

**Build fails:**
- Test build locally first
- Check Dockerfile and dependencies

**Image not found:**
- Check image visibility in GitHub package settings
- Login: `echo $GITHUB_TOKEN | docker login ghcr.io -u {username} --password-stdin`

## Resources

- [Release Guide](RELEASE_GUIDE.md) - How to create releases with version tags
- [GitHub Container Registry Docs](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
