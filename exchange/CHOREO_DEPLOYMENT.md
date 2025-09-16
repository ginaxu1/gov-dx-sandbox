# Choreo Deployment Guide

This guide explains how to deploy the exchange services to WSO2 Choreo.

## Problem Solved

The original Dockerfiles used `COPY ../shared/ /app/shared/` which works for local Docker Compose builds but fails in Choreo because:
- **Local builds**: Docker Compose sets build context to `/exchange/`, so `../shared/` resolves to `/shared/`
- **Choreo builds**: Build context is the service directory (e.g., `/exchange/policy-decision-point/`), so `../shared/` tries to go outside the build context

## Solution

Each service now has a `prepare-for-choreo.sh` script that copies the shared packages into the service directory before building.

## Services

### Policy Decision Point
- **Location**: `/exchange/policy-decision-point/`
- **Port**: 8082
- **Preparation Script**: `./prepare-for-choreo.sh`

### Consent Engine
- **Location**: `/exchange/consent-engine/`
- **Port**: 8081
- **Preparation Script**: `./prepare-for-choreo.sh`

## Deployment Steps

### 1. Prepare the Service

Before deploying to Choreo, run the preparation script:

```bash
cd /path/to/exchange/policy-decision-point
./prepare-for-choreo.sh
```

This will:
- Copy the shared packages into the service directory
- Make the service ready for Choreo deployment

### 2. Deploy to Choreo

1. Go to your Choreo console
2. Create a new component
3. Select "Dockerfile" as the build method
4. Point to the service directory (e.g., `/exchange/policy-decision-point/`)
5. Choreo will use the Dockerfile to build the image

### 3. Environment Variables

Each service uses different environment variable approaches:

#### Policy Decision Point & Consent Engine
- Uses command-line flags: `-env=production -port=8082`
- Environment variables: `ENVIRONMENT`, `PORT`, `LOG_LEVEL`, etc.

#### Orchestration Engine
- Uses `os.Getenv("PORT")` for port configuration
- Loads configuration from `config.json`

## Build Arguments

The Dockerfiles expect these build arguments (Choreo will provide defaults):
- `BUILD_VERSION`: Version of the build
- `BUILD_TIME`: Build timestamp
- `GIT_COMMIT`: Git commit hash

## Health Checks

All services include health check endpoints:
- Policy Decision Point: `http://localhost:8082/health`
- Consent Engine: `http://localhost:8081/health`
- Orchestration Engine: `http://localhost:4000/health`

## Troubleshooting

### Build Fails with "shared directory not found"
- Make sure you ran `./prepare-for-choreo.sh` before deploying
- Verify the shared packages are copied to the service directory

### Service Won't Start
- Check the environment variables in Choreo
- Verify the port configuration matches the service expectations

### Missing Dependencies
- Ensure all shared packages are properly copied
- Check that the `go.mod` replace directives are correct

## Local Development

For local development, continue using Docker Compose:

```bash
cd /path/to/exchange
docker compose up --build
```

This uses the original Dockerfiles with `../shared/` paths which work in the Docker Compose context.
