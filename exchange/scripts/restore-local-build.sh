#!/bin/bash
# Script to restore go.mod files for local development

echo "Restoring go.mod files for local development..."

# Function to restore go.mod for local build
restore_go_mod_for_local() {
    local service_dir="$1"
    local go_mod_file="$service_dir/go.mod"
    
    if [ -f "$go_mod_file" ]; then
        echo "Restoring $go_mod_file for local development..."
        
        # Replace Docker paths with local paths
        sed -i.bak 's|replace github.com/gov-dx-sandbox/exchange/shared/config => /app/shared/config|replace github.com/gov-dx-sandbox/exchange/shared/config => ../shared/config|g' "$go_mod_file"
        sed -i.bak 's|replace github.com/gov-dx-sandbox/exchange/shared/constants => /app/shared/constants|replace github.com/gov-dx-sandbox/exchange/shared/constants => ../shared/constants|g' "$go_mod_file"
        sed -i.bak 's|replace github.com/gov-dx-sandbox/exchange/shared/utils => /app/shared/utils|replace github.com/gov-dx-sandbox/exchange/shared/utils => ../shared/utils|g' "$go_mod_file"
        sed -i.bak 's|replace github.com/gov-dx-sandbox/exchange/shared/handlers => /app/shared/handlers|replace github.com/gov-dx-sandbox/exchange/shared/handlers => ../shared/handlers|g' "$go_mod_file"
        
        # Remove backup files
        rm -f "$go_mod_file.bak"
        
        echo "✅ Restored $go_mod_file"
    else
        echo "⚠️  $go_mod_file not found"
    fi
}

# Restore all service go.mod files
restore_go_mod_for_local "policy-decision-point"
restore_go_mod_for_local "consent-engine"
restore_go_mod_for_local "orchestration-engine-go"

echo "✅ All go.mod files restored for local development"
