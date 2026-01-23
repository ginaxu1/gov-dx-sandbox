# Release Guide

Releases are automated via GitHub Actions when a semantic version tag is pushed (e.g., `<VERSION>`).

## How to Release

```bash
# 1. Checkout main
git checkout main && git pull

# 2. Create & Push Tag
git tag <VERSION>
git push origin <VERSION>
```

This triggers the **Release** workflow which:
1.  Builds Docker images for all **8 services** (Backend + Frontend).
2.  Pushes tags: `<VERSION>`, `<MAJOR.MINOR>`, `<MAJOR>`, `latest`, and `sha-<commit>`.
3.  Scans images with Trivy.
4.  Creates a GitHub Release with changelogs.

## Manual Release
Go to **Actions** → **Release - Build and Publish All Services** → **Run workflow** → Enter version (e.g., `<VERSION>`).

## Artifacts

All images are published to **ghcr.io/opendif/opendif-core/**:

| Category | Service | Image Name |
| :--- | :--- | :--- |
| **Backend** | Portal Backend | `portal-backend` |
| | Audit Service | `audit-service` |
| | Policy Decision Point | `policy-decision-point` |
| | Consent Engine | `consent-engine` |
| | Orchestration Engine | `orchestration-engine` |
| **Frontend** | Admin Portal | `admin-portal` |
| | Consent Portal | `consent-portal` |
| | Member Portal | `member-portal` |

## Verification
```bash
docker pull ghcr.io/opendif/opendif-core/portal-backend:<VERSION>
```
