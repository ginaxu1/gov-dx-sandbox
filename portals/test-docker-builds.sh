#!/bin/bash

# Test script for Docker builds of React portals
# This script builds and tests each portal's Docker image

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Find an available host port starting from requested port
find_free_port() {
    local port=$1
    while lsof -ti tcp:"${port}" >/dev/null 2>&1 || docker ps --format '{{.Ports}}' | grep -q ":${port}->"; do
        port=$((port + 1))
    done
    echo "${port}"
}

# Function to test a portal
test_portal() {
    local portal_name=$1
    local portal_dir=$2
    local port=$3
    
    echo -e "${YELLOW}Testing ${portal_name}...${NC}"
    
    # Check if Dockerfile exists
    if [ ! -f "${portal_dir}/Dockerfile" ]; then
        echo -e "${RED}✗ Dockerfile not found in ${portal_dir}${NC}"
        return 1
    fi
    
    # Check if nginx.conf exists
    if [ ! -f "${portal_dir}/nginx.conf" ]; then
        echo -e "${RED}✗ nginx.conf not found in ${portal_dir}${NC}"
        return 1
    fi
    
    # Build the Docker image
    echo -e "${YELLOW}Building Docker image for ${portal_name}...${NC}"
    if docker build -t "${portal_name}:test" "${portal_dir}"; then
        echo -e "${GREEN}✓ Successfully built ${portal_name}${NC}"
    else
        echo -e "${RED}✗ Failed to build ${portal_name}${NC}"
        return 1
    fi
    
    # Run the container in the background
    # Clean up any previous container with the same name
    docker rm -f "${portal_name}-test" >/dev/null 2>&1 || true

    local host_port
    host_port=$(find_free_port "${port}")
    if [ "${host_port}" != "${port}" ]; then
        echo -e "${YELLOW}Port ${port} busy, using ${host_port} instead...${NC}"
    fi

    echo -e "${YELLOW}Starting container for ${portal_name} on port ${host_port}...${NC}"
    CONTAINER_ID=$(docker run -d -p "${host_port}:80" --name "${portal_name}-test" "${portal_name}:test" 2>/dev/null || echo "")
    
    if [ -z "$CONTAINER_ID" ]; then
        echo -e "${RED}✗ Failed to start container${NC}"
        return 1
    fi
    
    # Wait for container to be ready
    echo -e "${YELLOW}Waiting for container to be ready...${NC}"
    sleep 3
    
    # Check if container is running
    if ! docker ps | grep -q "${portal_name}-test"; then
        echo -e "${RED}✗ Container is not running${NC}"
        docker logs "${portal_name}-test"
        docker rm -f "${portal_name}-test" 2>/dev/null || true
        return 1
    fi
    
    # Test health check endpoint
    echo -e "${YELLOW}Testing health check endpoint...${NC}"
    if curl -f -s "http://localhost:${host_port}/health" | grep -q "healthy"; then
        echo -e "${GREEN}✓ Health check passed${NC}"
    else
        echo -e "${RED}✗ Health check failed${NC}"
        docker rm -f "${portal_name}-test" 2>/dev/null || true
        return 1
    fi
    
    # Test main page
    echo -e "${YELLOW}Testing main page...${NC}"
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:${host_port}/")
    if [ "$HTTP_CODE" = "200" ]; then
        echo -e "${GREEN}✓ Main page returns 200${NC}"
    else
        echo -e "${RED}✗ Main page returned ${HTTP_CODE}${NC}"
        docker rm -f "${portal_name}-test" 2>/dev/null || true
        return 1
    fi
    
    # Test React Router (should redirect to index.html)
    echo -e "${YELLOW}Testing React Router (404 redirect)...${NC}"
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:${host_port}/nonexistent-route")
    if [ "$HTTP_CODE" = "200" ]; then
        echo -e "${GREEN}✓ React Router redirect works (404 -> index.html)${NC}"
    else
        echo -e "${RED}✗ React Router redirect failed (got ${HTTP_CODE})${NC}"
        docker rm -f "${portal_name}-test" 2>/dev/null || true
        return 1
    fi
    
    # Check if index.html is served
    echo -e "${YELLOW}Verifying index.html content...${NC}"
    if curl -s "http://localhost:${host_port}/" | grep -q "<!DOCTYPE html\|<html"; then
        echo -e "${GREEN}✓ index.html is being served${NC}"
    else
        echo -e "${RED}✗ index.html not found or invalid${NC}"
        docker rm -f "${portal_name}-test" 2>/dev/null || true
        return 1
    fi
    
    # Cleanup
    echo -e "${YELLOW}Cleaning up...${NC}"
    docker stop "${portal_name}-test" >/dev/null 2>&1 || true
    docker rm -f "${portal_name}-test" >/dev/null 2>&1 || true
    
    echo -e "${GREEN}✓ All tests passed for ${portal_name}${NC}\n"
    return 0
}

# Main execution
echo -e "${YELLOW}=== Testing Portal Docker Builds ===${NC}\n"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PORTALS_DIR="${SCRIPT_DIR}"

# Test each portal
FAILED=0

if [ -d "${PORTALS_DIR}/admin-portal" ]; then
    test_portal "admin-portal" "${PORTALS_DIR}/admin-portal" "8080" || FAILED=1
fi

if [ -d "${PORTALS_DIR}/member-portal" ]; then
    test_portal "member-portal" "${PORTALS_DIR}/member-portal" "8081" || FAILED=1
fi

if [ -d "${PORTALS_DIR}/consent-portal" ]; then
    test_portal "consent-portal" "${PORTALS_DIR}/consent-portal" "8082" || FAILED=1
fi

# Summary
echo -e "${YELLOW}=== Test Summary ===${NC}"
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All portal tests passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    exit 1
fi

