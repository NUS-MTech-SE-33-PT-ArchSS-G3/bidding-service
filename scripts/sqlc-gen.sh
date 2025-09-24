#!/bin/bash

set -e

go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

echo "Generating SQLC code..."
sqlc generate -f sqlc/sqlc.yaml
echo "SQLC code gen complete"