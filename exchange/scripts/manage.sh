#!/bin/bash
# Exchange Services Management Script

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

if [ $# -eq 0 ]; then
    echo "Exchange Services Management"
    echo "Usage: $0 <command> [options]"
    echo ""
    echo "Commands:"
    echo "  start-local     - Start services in local environment"
    echo "  start-prod      - Start services in production environment"
    echo "  stop            - Stop all services"
    echo "  status          - Check service status and health"
    echo "  logs [service]  - View logs (all services or specific service)"
    echo ""
    echo "Services: PDP (${PDP_URL}), CE (${CE_URL})"
    echo "Available: policy-decision-point, consent-engine"
    echo ""
    echo "Examples: $0 start-local | $0 logs policy-decision-point | $0 status"
    exit 0
fi

COMMAND=$1
SERVICE=${2:-""}

case $COMMAND in
    "start-local")
        echo "Starting Exchange Services (Local)..."
        check_docker
        docker compose --env-file .env.local up --build -d
        wait_for_services
        check_health
        show_endpoints "Local"
        ;;
        
    "start-prod")
        echo "Starting Exchange Services (Production)..."
        check_docker
        docker compose --env-file .env.production up --build -d
        wait_for_services
        check_health
        show_endpoints "Production"
        ;;
        
    "stop")
        echo "Stopping Exchange Services..."
        docker compose down
        echo "✅ Stopped"
        ;;
        
    "status")
        echo "Exchange Services Status"
        echo "========================"
        echo ""
        echo "Container Status:"
        docker compose ps
        echo ""
        echo "Health Status:"
        check_health
        echo ""
        echo "Service Information:"
        show_endpoints "Current"
        echo ""
        echo "Recent Logs (last 10 lines):"
        echo "--- PDP ---"
        docker compose logs --tail=10 policy-decision-point 2>/dev/null || echo "No logs available"
        echo ""
        echo "--- CE ---"
        docker compose logs --tail=10 consent-engine 2>/dev/null || echo "No logs available"
        ;;
        
    "logs")
        if [ -n "$SERVICE" ]; then
            echo "Viewing logs for $SERVICE..."
            docker compose logs -f "$SERVICE"
        else
            echo "Viewing logs for all services..."
            echo "Available services: policy-decision-point, consent-engine"
            echo "Usage: $0 logs [service-name]"
            echo ""
            docker compose logs -f
        fi
        ;;
        
    "help")
        $0
        ;;
        
    *)
        echo "❌ Unknown command: $COMMAND"
        echo ""
        $0
        exit 1
        ;;
esac
