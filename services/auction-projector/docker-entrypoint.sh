#!/bin/sh
set -e

# Generate config.json from template with environment variable substitution
if [ -f /app/config.template.json ]; then
  envsubst < /app/config.template.json > /app/config.json
fi

# Execute the main application
exec "$@"