#!/bin/bash
set -e

echo "Building xk6-ibmmq plugin..."

# Build the Docker image
docker build -t xk6-ibmmq-builder .

# Run the build container to create the k6 binary with ibmmq extension
docker run --rm \
  -v "$(pwd):/workspace" \
  -w /workspace \
  -e MQ_INSTALLATION_PATH=/opt/mqm \
  -e LD_LIBRARY_PATH=/opt/mqm/lib64:/opt/mqm/lib \
  -e CGO_ENABLED=1 \
  xk6-ibmmq-builder \
  bash -c "
    echo 'Building k6 with IBM MQ extension...' &&
    xk6 build --with github.com/controllan/xk6-ibmmq=. --output k6
  "

echo "Build complete! Binary created at: ./k6"
