#!/bin/bash

echo "Removing old generated client..."
rm -r ../generated-client/bid-query/axios

npx @openapitools/openapi-generator-cli generate \
  -i ../openapi/bid-query/openapi.yaml \
  -g typescript-axios \
  -o ../generated-client/bid-query/axios \
  --additional-properties=npmName=@kei/bid-query-api-client,npmVersion=1.0.0,providedInRoot=true
