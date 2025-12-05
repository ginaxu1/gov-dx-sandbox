# Admin Portal

A React-based administrative dashboard for OpenDIF administrators to manage members, schemas, and system configurations.

## Overview

The Admin Portal provides a user interface for administrators to:
- Manage OpenDIF members (onboard/offboard)
- Review and approve schema submissions
- Monitor system health and audit logs
- Configure system-wide settings

**Technology**: React + TypeScript + TailwindCSS + Vite

## Features

- **Member Management** - Create, update, and delete member organizations
- **Schema Governance** - Review, approve, or reject data schemas submitted by members
- **Audit Log Viewer** - View and filter system audit logs
- **Dashboard** - High-level overview of system statistics

## Quick Start

### Prerequisites

- Node.js 18+
- npm 9+

### Run the Application

```bash
# Install dependencies
npm install

# Run in development mode
npm run dev
```

The application will be available at `http://localhost:5173` (or configured port).

## Configuration

### Environment Variables

Create a `.env` file based on `.env.template`:

```bash
VITE_API_BASE_URL=http://localhost:3000/api/v1  # Portal Backend API URL
VITE_AUTH_CLIENT_ID=your_client_id             # IdP Client ID
VITE_AUTH_ISSUER=your_issuer_url               # IdP Issuer URL
```

## Testing

```bash
# Run linting
npm run lint

# Run unit tests (if configured)
npm run test
```

## Docker

```bash
# Build image
docker build -t admin-portal .

# Run container
docker run -p 5173:80 \
  -e VITE_API_BASE_URL=http://localhost:3000/api/v1 \
  admin-portal
```
