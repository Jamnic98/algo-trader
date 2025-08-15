#!/usr/bin/env bash
set -e  # Exit on any error

echo "====================================="
echo " Running Dev Checks"
echo "====================================="

echo ""
echo "Formatting code with Black..."
uv run black .

echo ""
echo "Linting with Ruff..."
uv run ruff check --fix .

echo ""
echo "Type checking with mypy..."
uv run mypy .

echo ""
echo "Running tests with pytest..."
uv run pytest

echo ""
echo "All checks passed successfully!"
