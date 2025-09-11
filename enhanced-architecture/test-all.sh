#!/bin/bash

echo "🧪 Running Enhanced API Gateway Tests..."

cd api-gateway

# Run tests with coverage
echo "📊 Running tests with coverage..."
go test -v -race -coverprofile=coverage.out ./...

# Generate coverage report
if [ -f coverage.out ]; then
    echo "📈 Generating coverage report..."
    go tool cover -html=coverage.out -o coverage.html
    echo "✅ Coverage report generated: coverage.html"
fi

# Run linting (if available)
if command -v golangci-lint &> /dev/null; then
    echo "🔍 Running linter..."
    golangci-lint run
else
    echo "⚠️  golangci-lint not found, skipping linting"
fi

echo "✅ All tests completed"
