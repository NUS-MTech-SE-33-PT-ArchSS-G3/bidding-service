#!/bin/bash

echo "Removing old generated client..."
rm -r ../generated-client/bid-command/axios

npx @openapitools/openapi-generator-cli generate \
  -i ../openapi/bid-command/openapi.yaml \
  -g typescript-axios \
  -o ../generated-client/bid-command/axios \
  --additional-properties=npmName=@kei/bid-command-api-client,npmVersion=1.0.0,providedInRoot=true
