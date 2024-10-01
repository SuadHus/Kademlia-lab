#!/bin/bash

cd ../

COVERAGE_FILE="coverage.out"

# Run tests with coverage for all packages in the project
go test -p=1 -coverprofile="$COVERAGE_FILE" ./...

# Display coverage summary
go tool cover -func="$COVERAGE_FILE"

# Generate an HTML coverage report
go tool cover -html="$COVERAGE_FILE" -o coverage.html