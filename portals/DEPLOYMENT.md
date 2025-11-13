# Portal Deployment Guide

Essential commands to build and deploy the React portals (admin-portal, member-portal, consent-portal) using Docker.

## Prerequisites

- Docker installed and running

## Quick Start

### Admin Portal

```bash
# Build the Docker image
cd portals/admin-portal
docker build -t admin-portal:latest .

# Run the container
docker run -p 8080:80 admin-portal:latest

# The application will be available at http://localhost:8080
```

### Member Portal

```bash
# Build the Docker image
cd portals/member-portal
docker build -t member-portal:latest .

# Run the container
docker run -p 8081:80 member-portal:latest

# The application will be available at http://localhost:8081
```

### Consent Portal

```bash
# Build the Docker image
cd portals/consent-portal
docker build -t consent-portal:latest .

# Run the container
docker run -p 8082:80 consent-portal:latest

# The application will be available at http://localhost:8082
```

## Production Deployment

### Run in Background with Auto-restart

```bash
# Admin Portal
docker run -d -p 8080:80 --name admin-portal --restart unless-stopped admin-portal:latest

# Member Portal
docker run -d -p 8081:80 --name member-portal --restart unless-stopped member-portal:latest

# Consent Portal
docker run -d -p 8082:80 --name consent-portal --restart unless-stopped consent-portal:latest
```

## Container Management

```bash
# View running containers
docker ps

# View logs
docker logs admin-portal

# Stop container
docker stop admin-portal

# Remove container
docker rm admin-portal

# Remove image
docker rmi admin-portal:latest
```

## Health Check

```bash
curl http://localhost:8080/health
# Expected: healthy
```

## Quick Reference

| Portal | Port | Build | Run |
|--------|------|-------|-----|
| Admin Portal | 8080 | `docker build -t admin-portal:latest portals/admin-portal` | `docker run -p 8080:80 admin-portal:latest` |
| Member Portal | 8081 | `docker build -t member-portal:latest portals/member-portal` | `docker run -p 8081:80 member-portal:latest` |
| Consent Portal | 8082 | `docker build -t consent-portal:latest portals/consent-portal` | `docker run -p 8082:80 consent-portal:latest` |
