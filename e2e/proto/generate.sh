#!/bin/bash

set -e

echo "Generating proto files..."

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo "ERROR: protoc is not installed"
    echo "Install with: brew install protobuf"
    exit 1
fi

# Check and install protoc-gen-go if needed
if ! command -v protoc-gen-go &> /dev/null; then
    echo "protoc-gen-go not found. Installing..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    
    # Add $HOME/go/bin to PATH if not already there
    if [[ ":$PATH:" != *":$HOME/go/bin:"* ]]; then
        export PATH="$HOME/go/bin:$PATH"
    fi
fi

# Check and install protoc-gen-go-grpc if needed
if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "protoc-gen-go-grpc not found. Installing..."
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    
    # Add $HOME/go/bin to PATH if not already there
    if [[ ":$PATH:" != *":$HOME/go/bin:"* ]]; then
        export PATH="$HOME/go/bin:$PATH"
    fi
fi

# Ensure PATH includes go bin
if [[ ":$PATH:" != *":$HOME/go/bin:"* ]]; then
    export PATH="$HOME/go/bin:$PATH"
fi

protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       echo.proto

echo "Proto files generated!"
