#!/bin/bash
set -e

echo "Running xk6-ibmmq test against IBM MQ..."

# Use the compose project network
NETWORK_NAME="xk6-ibmmq_mq-network"

echo "Using network: $NETWORK_NAME"

# Run the test using the built k6 binary inside the builder container
docker run --rm \
  --network "$NETWORK_NAME" \
  -v "$(pwd):/workspace" \
  -w /workspace \
  -e MQ_INSTALLATION_PATH=/opt/mqm \
  -e LD_LIBRARY_PATH=/opt/mqm/lib64:/opt/mqm/lib \
  -e MQTRACE=1 \
  xk6-ibmmq-builder \
  bash -c "./k6 run minimal-test.js 2>&1 | head -100"

echo "Test completed!"
