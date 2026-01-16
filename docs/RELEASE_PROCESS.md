# Release Process

This document describes the process for creating a new release of the OpenDIF Core platform.

## Overview

The release workflow automatically builds and publishes Docker images for all services to GitHub Container Registry (GHCR) when a new version tag is pushed. The workflow also performs security scanning and creates a GitHub Release with detailed information about the published images.

## Release Components

The following services are built and published as part of each release:

- **Portal Backend** - Backend service for the data exchange portal
- **Audit Service** - Service for tracking and logging system events
- **Policy Decision Point** - Service for authorization decisions
- **Consent Engine** - Service for managing data consent workflows
- **Orchestration Engine** - Service for coordinating data exchange workflows

## Prerequisites

Before creating a release, ensure:

1. ✅ All changes are merged to the `main` branch
2. ✅ All tests are passing
3. ✅ You have write permissions to create tags and releases
4. ✅ The CHANGELOG is updated with release notes (if applicable)

## Release Methods

### Method 1: Tag-Based Release (Recommended)

Create and push a semantic version tag to trigger the release workflow automatically.

#### Step 1: Create a Version Tag

```bash
# Format: v<major>.<minor>.<patch>
# Example: v0.1.0, v1.0.0, v2.1.3

git tag v0.1.0
```

#### Step 2: Push the Tag

```bash
git push origin v0.1.0
```

#### Step 3: Monitor the Workflow

1. Go to the [Actions tab](../../actions) in GitHub
2. Watch the "Release - Build and Publish All Services" workflow
3. The workflow will:
   - Build Docker images for all 5 services
   - Tag images with version number (e.g., `0.1.0`) and `latest`
   - Push images to GHCR at `ghcr.io/opendif/opendif-core/<service>:<version>`
   - Run Trivy security scans on all images
   - Upload security scan results to GitHub Security tab
   - Create a GitHub Release with docker pull commands

### Method 2: Manual Workflow Dispatch

Trigger the release workflow manually from the GitHub UI without creating a git tag.

#### Step 1: Navigate to Actions

1. Go to the [Actions tab](../../actions)
2. Select "Release - Build and Publish All Services" from the workflows list

#### Step 2: Run the Workflow

1. Click "Run workflow" button
2. Enter the version (e.g., `v0.1.0`)
3. Click "Run workflow"

**Note:** Manual releases do not create a GitHub Release automatically. You'll need to create it manually if needed.

## Version Naming Convention

Follow [Semantic Versioning](https://semver.org/):

```
v<MAJOR>.<MINOR>.<PATCH>

Where:
- MAJOR: Incompatible API changes
- MINOR: Backwards-compatible functionality additions
- PATCH: Backwards-compatible bug fixes
```

**Examples:**

- `v0.1.0` - Initial beta release
- `v1.0.0` - First stable release
- `v1.1.0` - New features added
- `v1.1.1` - Bug fixes

## Docker Image Tags

Each release creates two tags per service:

1. **Version tag** (e.g., `0.1.0`) - Specific version, 'v' prefix is stripped
2. **Latest tag** (`latest`) - Always points to the most recent release

### Image Naming Format

```
ghcr.io/opendif/opendif-core/<service-name>:<tag>
```

**Examples:**

```bash
# Portal Backend
ghcr.io/opendif/opendif-core/portal-backend:0.1.0
ghcr.io/opendif/opendif-core/portal-backend:latest

# Audit Service
ghcr.io/opendif/opendif-core/audit-service:0.1.0
ghcr.io/opendif/opendif-core/audit-service:latest
```

## Pulling Released Images

After a successful release, pull images using:

```bash
# Pull specific version
docker pull ghcr.io/opendif/opendif-core/portal-backend:0.1.0
docker pull ghcr.io/opendif/opendif-core/audit-service:0.1.0
docker pull ghcr.io/opendif/opendif-core/policy-decision-point:0.1.0
docker pull ghcr.io/opendif/opendif-core/consent-engine:0.1.0
docker pull ghcr.io/opendif/opendif-core/orchestration-engine:0.1.0

# Or pull latest
docker pull ghcr.io/opendif/opendif-core/portal-backend:latest
```

## Security Scanning

Every release automatically runs Trivy vulnerability scanning on all images:

- **Severity Levels Checked:** CRITICAL, HIGH
- **Results Location:** GitHub Security tab → Code scanning alerts
- **Scan Reports:** Available as SARIF files in workflow artifacts

### Viewing Security Scan Results

1. Go to the **Security** tab in GitHub
2. Click **Code scanning**
3. Filter by workflow run to see results for a specific release

**Note:** Security scans may fail the workflow if critical vulnerabilities are found, but the workflow is configured to continue and publish images anyway (`continue-on-error: true`).

## Verifying a Release

After the workflow completes:

1. **Check GitHub Release**

   - Go to [Releases](../../releases)
   - Verify the new release is created with correct version
   - Review the release notes and docker pull commands

2. **Verify Images in GHCR**

   - Go to [Packages](../../packages)
   - Check all 5 service packages are published
   - Verify version tags are present

3. **Test Image Pull**

   ```bash
   docker pull ghcr.io/opendif/opendif-core/portal-backend:0.1.0
   ```

4. **Check Security Scans**
   - Go to **Security** → **Code scanning**
   - Review any detected vulnerabilities

## Troubleshooting

### Workflow Fails During Build

**Symptoms:** Build step fails for one or more services

**Solutions:**

- Check Dockerfile syntax and build context
- Ensure all source files are committed
- Verify build arguments are correct
- Check the build logs in the failed workflow run

### Images Not Appearing in GHCR

**Symptoms:** Workflow succeeds but images aren't in packages

**Solutions:**

- Verify you have `packages: write` permission
- Check that `GITHUB_TOKEN` has proper scopes
- Ensure the repository allows package creation

### Tag Already Exists Error

**Symptoms:** Cannot push tag because it already exists

**Solutions:**

```bash
# Delete local tag
git tag -d v0.1.0

# Delete remote tag (if you have permissions)
git push origin :refs/tags/v0.1.0

# Create and push new tag
git tag v0.1.0
git push origin v0.1.0
```

### Security Scan Failures

**Symptoms:** Trivy scan step shows failures

**Solutions:**

- Review vulnerability details in the Security tab
- Update base images in Dockerfiles to patched versions
- Consider security fixes in the next release
- Note: Scans are set to `continue-on-error`, so they won't block releases

## Best Practices

1. **Test Before Release**

   - Run all tests locally
   - Test Docker builds locally
   - Review all pending changes

2. **Version Incrementing**

   - Follow semantic versioning strictly
   - Document breaking changes in release notes
   - Use pre-release tags for beta versions (e.g., `v1.0.0-beta.1`)

3. **Release Notes**

   - Update CHANGELOG.md before tagging
   - Include notable changes, bug fixes, and breaking changes
   - Link to relevant issues and PRs

4. **Timing**

   - Schedule releases during low-traffic periods
   - Announce releases to the team
   - Coordinate with deployment schedules

5. **Security**
   - Review and address security scan results
   - Keep dependencies updated
   - Monitor for CVEs in base images

## Rollback Process

If a release needs to be rolled back:

1. **Do NOT delete the release or images** (breaks reproducibility)
2. **Create a new release** with the previous stable version
3. **Update deployment configurations** to point to the previous version:
   ```bash
   docker pull ghcr.io/opendif/opendif-core/portal-backend:0.0.9  # previous stable
   ```
4. **Document the rollback** in the CHANGELOG or release notes

## Support

For questions or issues with the release process:

- Open an issue in the repository
- Contact the platform maintainers
- Review workflow logs in the Actions tab

## Related Documentation

- [Contributing Guidelines](../CONTRIBUTING.md)
- [Makefile Usage](MAKEFILE_USAGE.md)
- [Local Development Setup](LOCAL_CODE_QUALITY_SETUP.md)
