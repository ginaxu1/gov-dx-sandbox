#!/bin/bash
# This script starts the Minikube cluster for local development.
# Ensure that Docker is running before executing this script.
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

if [ $? -ne 0 ]; then
  echo "Failed to start Minikube. Please check your Minikube installation."
  exit 1
fi

echo "Minikube started successfully. You can now deploy your applications."

# Set Minikube Docker environment
eval $(minikube docker-env)

# Build the Docker images for the provider wrappers
bal build --cloud="docker" provider-wrappers/dmt/

if [ $? -ne 0 ]; then
  echo "Failed to build the Docker images for provider wrappers."
  exit 1
fi
echo "Docker images for provider wrappers built successfully."


# Optionally, you can deploy your applications here
# Uncomment the following line to deploy a sample application
kubectl apply -f provider-wrappers/dmt/k8s/deployment.yaml

# Check if the deployment was successful
if [ $? -ne 0 ]; then
  echo "Failed to deploy the application. Please check your Kubernetes configuration."
  exit 1
fi

# Print the status of the Minikube cluster
minikube service provider-wrapper-dmt-service --url