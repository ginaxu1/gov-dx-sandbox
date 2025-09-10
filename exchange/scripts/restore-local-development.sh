#!/bin/bash
# Script to restore local development setup after Choreo deployment
# This should be run after Choreo deployment to restore local development

echo "Restoring local development setup..."

# Function to restore a service for local development
restore_service_for_local() {
    local service_dir="$1"
    local service_name="$2"
    
    echo "Restoring $service_name for local development..."
    
    # Remove shared directory from service (it's duplicated)
    if [ -d "$service_dir/shared" ]; then
        rm -rf "$service_dir/shared"
        echo "✅ Removed duplicate shared directory from $service_dir"
    fi
    
    # Update go.mod to use original shared directory for local development
    if [ -f "$service_dir/go.mod" ]; then
        echo "Updating $service_dir/go.mod for local development..."
        
        # Replace local paths with original shared directory paths
        sed -i.bak 's|replace github.com/gov-dx-sandbox/exchange/shared/config => ./shared/config|replace github.com/gov-dx-sandbox/exchange/shared/config => ../shared/config|g' "$service_dir/go.mod"
        sed -i.bak 's|replace github.com/gov-dx-sandbox/exchange/shared/constants => ./shared/constants|replace github.com/gov-dx-sandbox/exchange/shared/constants => ../shared/constants|g' "$service_dir/go.mod"
        sed -i.bak 's|replace github.com/gov-dx-sandbox/exchange/shared/utils => ./shared/utils|replace github.com/gov-dx-sandbox/exchange/shared/utils => ../shared/utils|g' "$service_dir/go.mod"
        sed -i.bak 's|replace github.com/gov-dx-sandbox/exchange/shared/handlers => ./shared/handlers|replace github.com/gov-dx-sandbox/exchange/shared/handlers => ../shared/handlers|g' "$service_dir/go.mod"
        
        # Remove backup files
        rm -f "$service_dir/go.mod.bak"
        
        echo "✅ Updated $service_dir/go.mod for local development"
    else
        echo "⚠️  $service_dir/go.mod not found"
    fi
    
    echo "✅ $service_name restored for local development"
}

# Restore all services
restore_service_for_local "consent-engine" "Consent Engine"
restore_service_for_local "policy-decision-point" "Policy Decision Point"
restore_service_for_local "orchestration-engine-go" "Orchestration Engine"

echo "✅ All services restored for local development"
echo ""
echo "📋 Local development is now ready:"
echo "1. Use 'make build' for Docker Compose builds"
echo "2. Use 'make test' to run tests"
echo "3. Use 'make start-local' to start services locally"
