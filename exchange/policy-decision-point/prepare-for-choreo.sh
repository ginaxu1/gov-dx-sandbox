#!/bin/bash

# Script to prepare the policy-decision-point service for Choreo deployment
# This copies the shared packages into the service directory

set -e

echo "Preparing policy-decision-point for Choreo deployment..."

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_DIR="$SCRIPT_DIR"
SHARED_DIR="$SERVICE_DIR/../shared"

# Check if shared directory exists
if [ ! -d "$SHARED_DIR" ]; then
    echo "ERROR: Shared directory not found at $SHARED_DIR"
    exit 1
fi

# Create shared directory in service directory if it doesn't exist
if [ ! -d "$SERVICE_DIR/shared" ]; then
    echo "Creating shared directory in service directory..."
    mkdir -p "$SERVICE_DIR/shared"
fi

# Copy shared packages
echo "Copying shared packages..."
cp -r "$SHARED_DIR"/* "$SERVICE_DIR/shared/"

echo "✅ Policy Decision Point prepared for Choreo deployment"
echo "   - Shared packages copied to: $SERVICE_DIR/shared/"
echo "   - You can now build and deploy to Choreo"
