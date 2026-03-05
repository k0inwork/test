#!/bin/bash

# Exit on any error
set -e

echo "Building core services..."
cd services/registry && go build -o registry main.go && cd ../..
cd services/task && go build -o task main.go && cd ../..
cd services/identity && go build -o identity main.go && cd ../..
cd services/network && go build -o network main.go && cd ../..

LOG_FILE="integration_test.log"
> $LOG_FILE

echo "Starting services..."
./services/registry/registry >> $LOG_FILE 2>&1 &
PIDS="$!"
sleep 2 # Let registry boot

./services/task/task >> $LOG_FILE 2>&1 &
PIDS="$PIDS $!"

./services/identity/identity >> $LOG_FILE 2>&1 &
PIDS="$PIDS $!"

./services/network/network >> $LOG_FILE 2>&1 &
PIDS="$PIDS $!"

# Wait for enough time for tasks to be registered and cron to fire at least once (@every 10s)
echo "Waiting 35 seconds for tasks to be registered and triggered..."
sleep 35

# Clean up processes
echo "Cleaning up..."
kill $PIDS 2>/dev/null || true

# Verification
echo "Verifying test results in log..."

SUCCESS=true

if grep -q "DUMMY_RECURRING_TEST_EXECUTED.*service.*identity" $LOG_FILE; then
    echo "✅ Identity task executed successfully."
else
    echo "❌ Identity task FAILED to execute."
    SUCCESS=false
fi

if grep -q "DUMMY_RECURRING_TEST_EXECUTED.*service.*network" $LOG_FILE; then
    echo "✅ Network task executed successfully."
else
    echo "❌ Network task FAILED to execute."
    SUCCESS=false
fi

# Clean up compiled binaries
rm services/registry/registry
rm services/task/task
rm services/identity/identity
rm services/network/network

if [ "$SUCCESS" = true ]; then
    echo "🎉 All dummy recurring tasks were successfully orchestrated!"
    exit 0
else
    echo "💥 Integration test failed. See $LOG_FILE for details."
    cat $LOG_FILE
    exit 1
fi
