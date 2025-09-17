#!/bin/bash

# Health check script for Gov DX API Server
# This script can be used by monitoring systems to check service health

set -e

# Configuration
SERVICE_URL="${SERVICE_URL:-http://localhost:3000}"
TIMEOUT="${TIMEOUT:-10}"
RETRIES="${RETRIES:-3}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Health check function
check_health() {
    local url="$1"
    local timeout="$2"
    
    log_info "Checking health at: $url"
    
    # Use curl with timeout
    response=$(curl -s -w "%{http_code}" -o /tmp/health_response.json --max-time "$timeout" "$url/health" 2>/dev/null || echo "000")
    
    if [ "$response" = "200" ]; then
        # Check if response contains expected fields
        if grep -q '"status":"healthy"' /tmp/health_response.json 2>/dev/null; then
            log_success "Service is healthy"
            return 0
        else
            log_error "Service responded but status is not healthy"
            return 1
        fi
    else
        log_error "Service returned HTTP $response"
        return 1
    fi
}

# Main health check with retries
main() {
    local retries="$RETRIES"
    local success=false
    
    while [ $retries -gt 0 ]; do
        if check_health "$SERVICE_URL" "$TIMEOUT"; then
            success=true
            break
        else
            retries=$((retries - 1))
            if [ $retries -gt 0 ]; then
                log_info "Retrying in 5 seconds... ($retries attempts left)"
                sleep 5
            fi
        fi
    done
    
    if [ "$success" = true ]; then
        log_success "Health check passed"
        exit 0
    else
        log_error "Health check failed after $RETRIES attempts"
        exit 1
    fi
}

# Cleanup function
cleanup() {
    rm -f /tmp/health_response.json
}

# Set up trap for cleanup
trap cleanup EXIT

# Handle command line arguments
case "${1:-check}" in
    "check")
        main
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  check  - Perform health check (default)"
        echo "  help   - Show this help message"
        echo ""
        echo "Environment Variables:"
        echo "  SERVICE_URL - Service URL (default: http://localhost:3000)"
        echo "  TIMEOUT     - Request timeout in seconds (default: 10)"
        echo "  RETRIES     - Number of retry attempts (default: 3)"
        ;;
    *)
        log_error "Unknown command: $1"
        echo "Use '$0 help' for usage information"
        exit 1
        ;;
esac
