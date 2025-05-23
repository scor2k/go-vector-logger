#!/bin/sh

# Run Vector container with TCP port 6000 published and vector.toml mounted

docker run --rm \
  --name vector \
  -p 6000:6000 \
  -v "$(pwd)/vector.toml:/etc/vector/vector.toml:ro" \
  -e VECTOR_LOG=info \
  timberio/vector:latest-alpine \
  -c /etc/vector/vector.toml
