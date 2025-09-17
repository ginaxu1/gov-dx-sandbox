#!/bin/bash

# Gov DX API Server - Choreo Deployment Script
# This script builds and deploys the API server to WSO2 Choreo

set -e

# Configuration
SERVICE_NAME="gov-dx-api-server"
SERVICE_VERSION="1.0.0"
REGISTRY_URL="choreo-registry"
NAMESPACE="choreo-system"
IMAGE_TAG="${SERVICE_NAME}:${SERVICE_VERSION}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if Docker is running
    if ! docker info > /dev/null 2>&1; then
        log_error "Docker is not running. Please start Docker and try again."
        exit 1
    fi
    
    # Check if kubectl is available
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed. Please install kubectl and try again."
        exit 1
    fi
    
    # Check if we can connect to the cluster
    if ! kubectl cluster-info > /dev/null 2>&1; then
        log_error "Cannot connect to Kubernetes cluster. Please check your kubeconfig."
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Build Docker image
build_image() {
    log_info "Building Docker image..."
    
    # Change to parent directory to build with proper context
    cd ..
    
    # Build the image
    docker build -f api-server-go/Dockerfile -t ${IMAGE_TAG} .
    
    if [ $? -eq 0 ]; then
        log_success "Docker image built successfully: ${IMAGE_TAG}"
    else
        log_error "Failed to build Docker image"
        exit 1
    fi
    
    # Change back to API server directory
    cd api-server-go
}

# Tag image for registry
tag_image() {
    log_info "Tagging image for registry..."
    
    docker tag ${IMAGE_TAG} ${REGISTRY_URL}/${IMAGE_TAG}
    
    if [ $? -eq 0 ]; then
        log_success "Image tagged successfully: ${REGISTRY_URL}/${IMAGE_TAG}"
    else
        log_error "Failed to tag image"
        exit 1
    fi
}

# Push image to registry
push_image() {
    log_info "Pushing image to registry..."
    
    docker push ${REGISTRY_URL}/${IMAGE_TAG}
    
    if [ $? -eq 0 ]; then
        log_success "Image pushed successfully to registry"
    else
        log_error "Failed to push image to registry"
        exit 1
    fi
}

# Deploy to Kubernetes
deploy_to_k8s() {
    log_info "Deploying to Kubernetes..."
    
    # Create namespace if it doesn't exist
    kubectl create namespace ${NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -
    
    # Apply the configuration
    kubectl apply -f choreo-config.yaml
    
    if [ $? -eq 0 ]; then
        log_success "Configuration applied successfully"
    else
        log_error "Failed to apply configuration"
        exit 1
    fi
    
    # Wait for deployment to be ready
    log_info "Waiting for deployment to be ready..."
    kubectl wait --for=condition=available --timeout=300s deployment/api-server -n ${NAMESPACE}
    
    if [ $? -eq 0 ]; then
        log_success "Deployment is ready"
    else
        log_error "Deployment failed to become ready"
        exit 1
    fi
}

# Verify deployment
verify_deployment() {
    log_info "Verifying deployment..."
    
    # Check if pods are running
    kubectl get pods -n ${NAMESPACE} -l app=api-server
    
    # Check service
    kubectl get service api-server-service -n ${NAMESPACE}
    
    # Check ingress
    kubectl get ingress api-server-ingress -n ${NAMESPACE}
    
    # Test health endpoint
    log_info "Testing health endpoint..."
    kubectl port-forward -n ${NAMESPACE} service/api-server-service 8080:80 &
    PORT_FORWARD_PID=$!
    
    sleep 5
    
    if curl -f http://localhost:8080/health > /dev/null 2>&1; then
        log_success "Health check passed"
    else
        log_warning "Health check failed - service may still be starting"
    fi
    
    # Clean up port forward
    kill $PORT_FORWARD_PID 2>/dev/null || true
}

# Cleanup function
cleanup() {
    log_info "Cleaning up..."
    # Add any cleanup logic here
}

# Main deployment function
main() {
    log_info "Starting deployment of ${SERVICE_NAME} v${SERVICE_VERSION}"
    
    # Set up trap for cleanup on exit
    trap cleanup EXIT
    
    # Run deployment steps
    check_prerequisites
    build_image
    tag_image
    push_image
    deploy_to_k8s
    verify_deployment
    
    log_success "Deployment completed successfully!"
    log_info "Service is available at: https://api.govdx-sandbox.gov"
    log_info "API Documentation: https://api.govdx-sandbox.gov/openapi.yaml"
}

# Handle command line arguments
case "${1:-deploy}" in
    "deploy")
        main
        ;;
    "build")
        check_prerequisites
        build_image
        ;;
    "push")
        check_prerequisites
        build_image
        tag_image
        push_image
        ;;
    "k8s")
        deploy_to_k8s
        verify_deployment
        ;;
    "verify")
        verify_deployment
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  deploy  - Full deployment (default)"
        echo "  build   - Build Docker image only"
        echo "  push    - Build, tag, and push image"
        echo "  k8s     - Deploy to Kubernetes only"
        echo "  verify  - Verify deployment"
        echo "  help    - Show this help message"
        ;;
    *)
        log_error "Unknown command: $1"
        echo "Use '$0 help' for usage information"
        exit 1
        ;;
esac
