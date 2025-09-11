#!/bin/bash
# Script to prepare services for Choreo deployment by copying shared directory

echo "Preparing services for Choreo deployment..."

# Function to prepare a service for Choreo
prepare_service_for_choreo() {
    local service_dir="$1"
    local service_name="$2"
    
    echo "Preparing $service_name for Choreo deployment..."
    
    # Create shared directory in service if it doesn't exist
    mkdir -p "$service_dir/shared"
    
    # Copy shared packages into service directory
    cp -r shared/* "$service_dir/shared/"
    
    # Update go.mod to use local shared directory
    if [ -f "$service_dir/go.mod" ]; then
        echo "Updating $service_dir/go.mod for Choreo deployment..."
        
        # Replace all paths with local shared directory paths for Choreo
        sed -i.bak 's|replace github.com/gov-dx-sandbox/exchange/shared/config => /app/shared/config|replace github.com/gov-dx-sandbox/exchange/shared/config => ./shared/config|g' "$service_dir/go.mod"
        sed -i.bak 's|replace github.com/gov-dx-sandbox/exchange/shared/constants => /app/shared/constants|replace github.com/gov-dx-sandbox/exchange/shared/constants => ./shared/constants|g' "$service_dir/go.mod"
        sed -i.bak 's|replace github.com/gov-dx-sandbox/exchange/shared/utils => /app/shared/utils|replace github.com/gov-dx-sandbox/exchange/shared/utils => ./shared/utils|g' "$service_dir/go.mod"
        sed -i.bak 's|replace github.com/gov-dx-sandbox/exchange/shared/handlers => /app/shared/handlers|replace github.com/gov-dx-sandbox/exchange/shared/handlers => ./shared/handlers|g' "$service_dir/go.mod"
        
        # Also replace local development paths with Choreo paths
        sed -i.bak 's|replace github.com/gov-dx-sandbox/exchange/shared/config => ../shared/config|replace github.com/gov-dx-sandbox/exchange/shared/config => ./shared/config|g' "$service_dir/go.mod"
        sed -i.bak 's|replace github.com/gov-dx-sandbox/exchange/shared/constants => ../shared/constants|replace github.com/gov-dx-sandbox/exchange/shared/constants => ./shared/constants|g' "$service_dir/go.mod"
        sed -i.bak 's|replace github.com/gov-dx-sandbox/exchange/shared/utils => ../shared/utils|replace github.com/gov-dx-sandbox/exchange/shared/utils => ./shared/utils|g' "$service_dir/go.mod"
        sed -i.bak 's|replace github.com/gov-dx-sandbox/exchange/shared/handlers => ../shared/handlers|replace github.com/gov-dx-sandbox/exchange/shared/handlers => ./shared/handlers|g' "$service_dir/go.mod"
        
        # Remove backup files
        rm -f "$service_dir/go.mod.bak"
        
        echo "✅ Updated $service_dir/go.mod for Choreo deployment"
    else
        echo "⚠️  $service_dir/go.mod not found"
    fi
    
    echo "✅ $service_name prepared for Choreo deployment"
}

# Prepare all services
prepare_service_for_choreo "consent-engine" "Consent Engine"
prepare_service_for_choreo "policy-decision-point" "Policy Decision Point"
prepare_service_for_choreo "orchestration-engine-go" "Orchestration Engine"

echo "✅ All services prepared for Choreo deployment"
echo ""
echo "📋 Next steps for Choreo deployment:"
echo "1. Use Dockerfile for each service"
echo "2. Set build context to the service directory (e.g., consent-engine/)"
echo "3. The shared directory is now included in each service directory"
