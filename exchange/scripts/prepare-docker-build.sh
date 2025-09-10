#!/bin/bash
# Script to prepare go.mod files for Docker/Choreo builds

echo "Preparing go.mod files for Docker/Choreo builds..."

# Function to update go.mod for Docker build
update_go_mod_for_docker() {
    local service_dir="$1"
    local go_mod_file="$service_dir/go.mod"
    
    if [ -f "$go_mod_file" ]; then
        echo "Updating $go_mod_file for Docker build..."
        
        # Replace local paths with Docker paths
        sed -i.bak 's|replace github.com/gov-dx-sandbox/exchange/shared/config => ../shared/config|replace github.com/gov-dx-sandbox/exchange/shared/config => /app/shared/config|g' "$go_mod_file"
        sed -i.bak 's|replace github.com/gov-dx-sandbox/exchange/shared/constants => ../shared/constants|replace github.com/gov-dx-sandbox/exchange/shared/constants => /app/shared/constants|g' "$go_mod_file"
        sed -i.bak 's|replace github.com/gov-dx-sandbox/exchange/shared/utils => ../shared/utils|replace github.com/gov-dx-sandbox/exchange/shared/utils => /app/shared/utils|g' "$go_mod_file"
        sed -i.bak 's|replace github.com/gov-dx-sandbox/exchange/shared/handlers => ../shared/handlers|replace github.com/gov-dx-sandbox/exchange/shared/handlers => /app/shared/handlers|g' "$go_mod_file"
        
        # Remove backup files
        rm -f "$go_mod_file.bak"
        
        echo "✅ Updated $go_mod_file"
    else
        echo "⚠️  $go_mod_file not found"
    fi
}

# Update all service go.mod files
update_go_mod_for_docker "policy-decision-point"
update_go_mod_for_docker "consent-engine"
update_go_mod_for_docker "orchestration-engine-go"

echo "✅ All go.mod files prepared for Docker/Choreo builds"
