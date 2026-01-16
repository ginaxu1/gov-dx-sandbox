# Release Guide

This guide covers how to create and publish releases for the OpenDIF Core platform.

## Quick Start

Creating a release is simple:

```bash
# Create and push a version tag
git tag v0.1.0
git push origin v0.1.0
```

This automatically builds and publishes Docker images for all services to GitHub Container Registry (GHCR).

## What Gets Released

Each release includes Docker images for all services:

- **portal-backend** - Portal backend API
- **audit-service** - Audit logging service
- **policy-decision-point** - Authorization service
- **consent-engine** - Consent management service
- **orchestration-engine** - Data exchange orchestration

## Version Naming

Follow [Semantic Versioning](https://semver.org/): `v<MAJOR>.<MINOR>.<PATCH>`

**Examples:**

- `v0.1.0` - Initial beta
- `v1.0.0` - First stable release
- `v1.1.0` - New features (backward compatible)
- `v1.1.1` - Bug fixes

**When to increment:**

- **MAJOR** - Breaking changes
- **MINOR** - New features (backward compatible)
- **PATCH** - Bug fixes

## How to Release

### Method 1: Tag-Based Release (Recommended)

Push a semantic version tag to trigger automatic release:

```bash
# Ensure you're on main and up to date
git checkout main
git pull origin main

# Create and push tag
git tag v0.1.0
git push origin v0.1.0
```

**What happens automatically:**

1. Builds all 5 service Docker images
2. Tags images as `0.1.0` (v prefix stripped) and `latest`
3. Pushes images to `ghcr.io/opendif/opendif-core/<service>:<version>`
4. Runs security scans (Trivy)
5. Creates GitHub Release with pull commands

### Method 2: Manual Dispatch

For testing or special cases, trigger manually:

1. Go to **Actions** → "Release - Build and Publish All Services"
2. Click **Run workflow**
3. Select branch: `main`
4. Enter version: `v0.1.0`
5. Click **Run workflow**

**Note:** Manual dispatch from `main` will push images. From feature branches, it runs in dry-run mode (build only, no push).

## Image Tags

Each service gets two tags:

```
ghcr.io/opendif/opendif-core/<service>:0.1.0    # Version tag
ghcr.io/opendif/opendif-core/<service>:latest    # Latest tag
```

**Pull images:**

```bash
docker pull ghcr.io/opendif/opendif-core/portal-backend:0.1.0
docker pull ghcr.io/opendif/opendif-core/audit-service:0.1.0
docker pull ghcr.io/opendif/opendif-core/policy-decision-point:0.1.0
docker pull ghcr.io/opendif/opendif-core/consent-engine:0.1.0
docker pull ghcr.io/opendif/opendif-core/orchestration-engine:0.1.0
```

## Testing the Workflow

Before merging release workflow changes:

1. Push your feature branch to GitHub
2. Manually trigger workflow on that branch
3. It runs in **dry-run mode**: builds images but doesn't push
4. Verify all builds complete successfully
5. Once confirmed, merge to main

## Best Practices

### Before Release

- ✅ All changes merged to `main`
- ✅ All tests passing
- ✅ Update CHANGELOG.md (if exists)
- ✅ Review pending changes and breaking changes
- ✅ Test Docker builds locally if making Dockerfile changes

### During Release

- Use clear version numbers following semver
- Tag releases during low-traffic periods
- Monitor the Actions workflow for completion
- Check GitHub Releases page for automatic release creation

### After Release

- Verify images appear in [Packages](../../packages)
- Test pulling at least one image
- Review security scan results in [Security](../../security/code-scanning) tab
- Announce release to team if needed

### Security

- Review Trivy scan results after each release
- Keep base images updated
- Address critical vulnerabilities promptly
- Security scans continue even if vulnerabilities found

## Verification

After release completes:

1. **Check GitHub Release:**

   - Visit [Releases](../../releases)
   - Verify release notes and version

2. **Verify Images:**

   - Go to [Packages](../../packages)
   - Check all 5 services are published

3. **Test Pull:**
   ```bash
   docker pull ghcr.io/opendif/opendif-core/portal-backend:0.1.0
   ```

## Troubleshooting

### Build Failures

Check workflow logs in Actions tab for:

- Dockerfile syntax errors
- Missing dependencies
- Build context issues

### Tag Already Exists

```bash
# Delete and recreate tag
git tag -d v0.1.0
git push origin :refs/tags/v0.1.0
git tag v0.1.0
git push origin v0.1.0
```

### Images Not Published

- Verify you have `packages: write` permission
- Ensure workflow completed successfully
- Check if dry-run mode was enabled (feature branch)

## Rollback

If a release has issues:

1. **Do NOT delete the release** (breaks reproducibility)
2. Create a new release with previous stable version
3. Update deployments to use previous version tag
4. Document the rollback reason

**Example:**

```bash
# Roll back to v0.0.9
docker pull ghcr.io/opendif/opendif-core/portal-backend:0.0.9
```

## Workflow Details

The release workflow:

- **Trigger:** Push tag matching `v*.*.*` or manual dispatch
- **Builds:** All 5 services in parallel
- **Registry:** GitHub Container Registry (ghcr.io)
- **Security:** Trivy scans for CRITICAL and HIGH vulnerabilities
- **Artifacts:** Docker images and security scan reports
- **Dry-run:** Enabled for non-main branches (build only, no push)

## Related Documentation

- [Contributing Guidelines](../CONTRIBUTING.md)
- [Makefile Usage](MAKEFILE_USAGE.md)
- [Local Development Setup](LOCAL_CODE_QUALITY_SETUP.md)

## Support

Questions or issues?

- Check [Actions](../../actions) workflow logs
- Review [existing releases](../../releases) for examples
- Open an issue in the repository
