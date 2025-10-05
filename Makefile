GO ?= go
CMD_DIR := cmd
BIN_DIR := bin

# space separated list
BINARIES := bid-command bid-query

.PHONY: all build run clean generate generate-client check-generate

all: build

build:
	go mod tidy
	mkdir -p $(BIN_DIR)
	@for b in $(BINARIES); do \
		echo "Building $$b..."; \
		cd $(CMD_DIR)/$$b && $(GO) build -o ../../$(BIN_DIR)/$$b; \
		cd - > /dev/null; \
	done

run:
	go mod tidy
	cd $(CMD_DIR)/$(CMD) && $(GO) run . -config ../../cmd/$(CMD)/config.json

runsmx:
	go mod tidy
	cd $(CMD_DIR)/$(CMD) && $(GO) run . -config ../../cmd/$(CMD)/configsmx.json

clean:
	rm -rf $(BIN_DIR)

generate:
	go generate ./...

generate-client:
	rm -r ./openapi/frontend/angular
	npx @openapitools/openapi-generator-cli generate \
	  -i ./openapi/control/openapi.yaml \
	  -g typescript-angular \
	  -o ./openapi/frontend/angular \
	  --additional-properties=npmName=@yourorg/control-api-client,npmVersion=1.0.0,providedInRoot=true

#check-generate:
#	go generate ./...
#	git diff --exit-code || (echo "Generated code out of date"; exit 1)