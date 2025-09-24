#!/bin/bash

echo "Removing old generated client..."
rm -r ../generated-client/axios

npx @openapitools/openapi-generator-cli generate \
  -i ../openapi/openapi.yaml \
  -g typescript-axios \
  -o ../generated-client/axios \
  --additional-properties=npmName=@kei/bidding-api-client,npmVersion=1.0.0,providedInRoot=true
