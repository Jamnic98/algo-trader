#!/usr/bin/env bash
set -e  # Exit on any error

echo "====================================="
echo " Running Dev Checks"
echo "====================================="

echo ""
echo "Formatting code with Black..."
black .

echo ""
echo "Linting with Ruff..."
ruff check --fix .

echo ""
echo "Type checking with mypy..."
mypy .

echo ""
echo "Running tests with pytest..."
pytest

echo ""
echo "All checks passed successfully!"
