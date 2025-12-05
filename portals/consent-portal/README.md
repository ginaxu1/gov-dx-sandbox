# Consent Portal

A citizen-facing React application that allows data owners to view, approve, or deny data access requests.

## Overview

The Consent Portal is the interface where citizens (data owners) interact with OpenDIF to manage their data consents. It is typically accessed via a redirect from a data consumer application when consent is required.

**Technology**: React + TypeScript + TailwindCSS + Vite

## Features

- **Consent Review** - View details of data access requests (who, what, why)
- **Approval/Denial** - Grant or deny access to requested data
- **Consent Management** - View and revoke previously granted consents
- **Secure Authentication** - Integration with IdP for user authentication

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

The application will be available at `http://localhost:5174` (or configured port).

## Configuration

### Environment Variables

Create a `.env` file based on `.env.template`:

```bash
VITE_API_BASE_URL=http://localhost:8081  # Consent Engine API URL
VITE_AUTH_CLIENT_ID=your_client_id       # IdP Client ID
VITE_AUTH_ISSUER=your_issuer_url         # IdP Issuer URL
```

## Testing Guide

### End-to-End Flow

1. **Start Backend Services**: Ensure Consent Engine is running on port 8081.
2. **Generate Consent Request**:
   - Use Postman or curl to create a consent request in Consent Engine.
   - Copy the `consent_id` from the response.
3. **Access Portal**:
   - Navigate to `http://localhost:5174/?consent={consent_id}`
   - Log in if required.
   - Review and act on the consent request.

## Docker

```bash
# Build image
docker build -t consent-portal .

# Run container
docker run -p 5174:80 \
  -e VITE_API_BASE_URL=http://localhost:8081 \
  consent-portal
```