# Member Portal

A React-based dashboard for OpenDIF members to manage their data schemas, applications, and integration settings.

## Overview

The Member Portal allows participating organizations (members) to:
- Define and submit data schemas
- Register applications to consume data
- Manage API keys and integration settings
- View usage analytics

**Technology**: React + TypeScript + TailwindCSS + Vite

## Features

- **Schema Management** - Create, edit, and submit data schemas for approval
- **Application Registry** - Register new applications and manage their credentials
- **Integration Hub** - Configure webhooks and other integration points
- **Analytics** - View data exchange metrics

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
docker build -t member-portal .

# Run container
docker run -p 5173:80 \
  -e VITE_API_BASE_URL=http://localhost:3000/api/v1 \
  member-portal
```
