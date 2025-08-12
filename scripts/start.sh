#!/bin/bash
# This script starts the Minikube cluster for local development.
# Ensure that Docker is running before executing this script.
# Check if Docker is running

# Define a function to perform the cleanup
cleanup() {
  echo "Caught Ctrl+C! Cleaning up Kubernetes resources..."
  # Delete all deployments
  kubectl delete deployment graphql-resolver-deployment
  kubectl delete deployment provider-wrapper-drp-deployment
  kubectl delete deployment provider-wrapper-dmt-deployment
  kubectl delete deployment mock-drp-deployment

  # Delete all services
  kubectl delete service graphql-resolver-service
  kubectl delete service provider-wrapper-drp-service
  kubectl delete service provider-wrapper-dmt-service
  kubectl delete service mock-drp-service

  echo "Cleanup complete. Exiting."
  exit 0
}

# Set the trap: when SIGINT is received, run the cleanup function
trap cleanup SIGINT

# Fail on any error
set -e

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
  echo "Docker is not running. Please start Docker and try again."
  exit 1
fi

# Check if Minikube is installed
if ! command -v minikube &> /dev/null; then
  echo "Minikube is not installed. Please install Minikube and try again."
  exit 1
fi

# Check if the Minikube profile exists
if minikube profile list | grep -q 'minikube'; then
  echo "Minikube profile 'minikube' already exists. Skipping profile creation."
else
  echo "Creating Minikube profile 'minikube'."
  minikube profile create minikube
  if [ $? -ne 0 ]; then
    echo "Failed to create Minikube profile. Please check your Minikube installation."
    exit 1
  fi
fi

# Start Minikube if not already running
if ! minikube status | grep -q 'Running'; then
  echo "Starting Minikube..."
  minikube start --driver=docker
else
  echo "Minikube is already running."
fi

# Check if Minikube started successfully
if [ $? -ne 0 ]; then
  echo "Failed to start Minikube. Please check your Minikube installation."
  exit 1
else
  echo "Minikube started successfully. You can now deploy your applications."
fi


# Set Minikube Docker environment
eval $(minikube docker-env)

echo "Starting parallel Docker builds..."

# Build the Docker images for the mocks
bal build --cloud="docker" mocks/mock-drp/ &

if [ $? -ne 0 ]; then
  echo "Failed to build the Docker images for mocks."
  exit 1
fi

# Build the Docker images for the provider wrappers
bal build --cloud="docker" provider-wrappers/dmt/ &
bal build --cloud="docker" provider-wrappers/drp/ &
docker build -t gov-dx-sandbox/graphql-resolver:v0.1.3 graphql-resolver/ &

wait
echo "All Docker builds completed successfully."

if [ $? -ne 0 ]; then
  echo "Failed to build the Docker images for provider wrappers."
  exit 1
else
  echo "Docker images for provider wrappers built successfully."
fi

# Deploy the mock services
kubectl apply -f mocks/mock-drp/k8s/deployment.yaml

# Deploy the provider wrappers
kubectl apply -f provider-wrappers/dmt/k8s/deployment.yaml
kubectl apply -f provider-wrappers/drp/k8s/deployment.yaml

# Deploy the GraphQL resolver
kubectl apply -f graphql-resolver/k8s/deployment.yaml

# Check if the deployment was successful
if [ $? -ne 0 ]; then
  echo "Failed to deploy the application. Please check your Kubernetes configuration."
  exit 1
fi

# Print the status of the Minikube cluster
minikube tunnel

# The script will only reach this point if minikube tunnel exits on its own
echo "Tunnel stopped, cleaning up."
cleanup