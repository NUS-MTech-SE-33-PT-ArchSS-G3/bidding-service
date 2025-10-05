#!/bin/bash

echo "Removing old generated client..."
rm -r ../generated-client/axios/bid-command

npx @openapitools/openapi-generator-cli generate \
  -i ../openapi/openapi-bid-command/openapi.yaml \
  -g typescript-axios \
  -o ../generated-client/bid-command/axios \
  --additional-properties=npmName=@kei/bidding-command-api-client,npmVersion=1.0.0,providedInRoot=true
