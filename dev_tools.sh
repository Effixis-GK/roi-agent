#!/bin/bash

BASE_DIR="/Users/taktakeu/Local/GitHub/desktop_app"

case "$1" in
    "status")
        echo "=== Agent Status ==="
        if pgrep -f "monitor" > /dev/null; then
            echo "Agent: RUNNING"
            "$BASE_DIR/agent/monitor" status 2>/dev/null || echo "Status check failed"
        else
            echo "Agent: NOT RUNNING"
        fi
        
        echo -e "\n=== Web UI Status ==="
        if pgrep -f "python.*app.py" > /dev/null; then
            echo "Web UI: RUNNING (http://localhost:5002)"
        else
            echo "Web UI: NOT RUNNING"
        fi
        ;;
        
    "logs")
        echo "=== Recent Agent Logs ==="
        tail -n 20 "$BASE_DIR/logs/agent.log" 2>/dev/null || echo "No agent logs found"
        ;;
        
    "permissions")
        echo "=== Checking macOS Permissions ==="
        "$BASE_DIR/agent/monitor" check-permissions 2>/dev/null || echo "Cannot check permissions - build agent first"
        ;;
        
    "clean")
        echo "=== Cleaning temporary files ==="
        rm -f "$BASE_DIR/logs/"*.log
        rm -f "$BASE_DIR/data/"*.json
        echo "Cleaned logs and data files"
        ;;
        
    "build")
        echo "=== Building all components ==="
        cd "$BASE_DIR"
        ./build_agent.sh
        echo "Build complete"
        ;;
        
    *)
        echo "macOS Application Monitor - Development Tools"
        echo ""
        echo "Usage: $0 {status|logs|permissions|clean|build}"
        echo ""
        echo "Commands:"
        echo "  status      - Show running status of agent and web UI"
        echo "  logs        - Show recent agent logs"
        echo "  permissions - Check macOS accessibility permissions"
        echo "  clean       - Clean temporary files and logs"
        echo "  build       - Build all components"
        ;;
esac
