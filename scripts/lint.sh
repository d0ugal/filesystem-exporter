#!/bin/bash

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}🔍 Running golangci-lint using official container...${NC}"

# Run golangci-lint using the official container
echo -e "${YELLOW}🔧 Running golangci-lint...${NC}"
if docker run --rm \
    -v "$(pwd):/app" \
    -w /app \
    golangci/golangci-lint:latest \
    golangci-lint run; then
    echo -e "${GREEN}✅ Linting passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Linting failed!${NC}"
    exit 1
fi
