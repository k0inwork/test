#!/bin/bash

# Configuration
SSH_CONFIG="/tmp/.ssh/config"
TARGET_HOST="10.0.111.182"
LOCAL_PORT="${LOCAL_PORT:-8080}"
REMOTE_PORT="80"
WAIT_TIME=5

ADMIN_USERNAME="${ADMIN_USERNAME:-admin}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-adminadmin}"

# Ensure SSH key permissions are correct
chmod 600 /tmp/.ssh/id_rsa_tablet

# Function to clean up the SSH tunnel on exit
cleanup() {
    if [ -n "$TUNNEL_PID" ]; then
        echo "Killing SSH tunnel (PID: $TUNNEL_PID)..."
        kill "$TUNNEL_PID" 2>/dev/null
    fi
}

# Trap EXIT and ERR signals to ensure cleanup
trap cleanup EXIT ERR

# Start the SSH tunnel
echo "Starting SSH tunnel to $TARGET_HOST ($LOCAL_PORT:$REMOTE_PORT) via bridge..."
ssh -F "$SSH_CONFIG" bridge -L "$LOCAL_PORT:$TARGET_HOST:$REMOTE_PORT" -N -o StrictHostKeyChecking=no &
TUNNEL_PID=$!

echo "SSH tunnel started with PID: $TUNNEL_PID"
echo "Waiting $WAIT_TIME seconds for the tunnel to establish..."
sleep "$WAIT_TIME"

# Check if the tunnel process is still running
if ! kill -0 "$TUNNEL_PID" 2>/dev/null; then
    echo "Error: SSH tunnel process died unexpectedly."
    exit 1
fi

# Run the scraping script
echo "Running scrape_legacy.py..."
# Assumes .venv is active or requests is available
python3 scrape_legacy.py --username "$ADMIN_USERNAME" --password "$ADMIN_PASSWORD" --url "http://localhost:$LOCAL_PORT"
SCRAPE_EXIT_CODE=$?

# If the script failed, retry once more without the 5s delay at the end
if [ $SCRAPE_EXIT_CODE -ne 0 ]; then
    echo "Scrape failed, retrying..."
    python3 scrape_legacy.py --username "$ADMIN_USERNAME" --password "$ADMIN_PASSWORD" --url "http://localhost:$LOCAL_PORT"
    SCRAPE_EXIT_CODE=$?
fi

if [ $SCRAPE_EXIT_CODE -ne 0 ]; then
    echo "Error: scrape_legacy.py failed with exit code $SCRAPE_EXIT_CODE"
    exit $SCRAPE_EXIT_CODE
fi

echo "Scraping completed successfully."