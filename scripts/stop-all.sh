#!/bin/bash

# Stop all running services
pkill -f main.go || true
echo "All services have been stopped."