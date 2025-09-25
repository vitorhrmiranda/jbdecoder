#!/bin/bash

set -e

echo "Installing jbdecoder..."

go install github.com/vitorhrmiranda/jbdecoder/cmd/cli@latest

# Find the GOPATH/bin path
GOBIN=$(go env GOPATH)/bin

# Check if binary was installed
if [ ! -f "$GOBIN/cli" ]; then
    echo "Error: Could not find installed binary at $GOBIN/cli"
    exit 1
fi

# Rename binary from 'cli' to 'jbdecoder'
mv "$GOBIN/cli" "$GOBIN/jbdecoder"

echo "✅ jbdecoder successfully installed at $GOBIN/jbdecoder"