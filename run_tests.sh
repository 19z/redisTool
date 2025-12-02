#!/bin/bash

echo "Running Redis Tool Tests..."
echo ""

cd "$(dirname "$0")"

echo "Installing dependencies..."
go mod tidy

echo ""
echo "Running tests with coverage..."
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

if [ $? -eq 0 ]; then
    echo ""
    echo "Tests passed! Generating coverage report..."
    go tool cover -html=coverage.out -o coverage.html
    echo "Coverage report generated: coverage.html"
    echo ""
    go tool cover -func=coverage.out
else
    echo ""
    echo "Tests failed!"
    exit 1
fi
