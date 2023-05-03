GO_BIN=$(shell pwd)/.bin/go

SHELL:=env PATH=$(GO_BIN):$(PATH) $(SHELL)

GOCI_LINT_V?=v1.51.0

.PHONY: all
all: lint test build

.PHONY: build
build:
	go build -a -o vervet ./cmd/vervet

# Run go mod tidy yourself

#----------------------------------------------------------------------------------
# Check for updates to packages in remote OSS repositories and update go.mod AND
# go.sum to match changes. Then download the all the dependencies
# This catches when your app has colliding versions of packages during updates
#----------------------------------------------------------------------------------
.PHONY: update-deps
update-deps:
	go get -d -u ./...

# go mod download yourself if you don't need to update

.PHONY: lint
lint:
	golangci-lint run -v ./...
	(cd versionware/example; golangci-lint run -v ./...)

.PHONY: lint-docker
lint-docker:
	docker run --rm -v $(shell pwd):/vervet -w /vervet golangci/golangci-lint:v1.43 golangci-lint run -v ./...

#----------------------------------------------------------------------------------
#  Ignores the test cache and forces a full test suite execution
#----------------------------------------------------------------------------------
.PHONY: test
test:
	go test ./... -count=1
	(cd versionware/example; go generate . && go test ./... -count=1)

.PHONY: test-coverage
test-coverage:
	go test ./... -count=1 -coverprofile=covfile
	go tool cover -html=covfile
	rm -f covfile

.PHONY: clean
clean:
	$(RM) vervet

.PHONY: install-tools
install-tools: 
ifndef CI
	mkdir -p ${GO_BIN}
	curl -sSfL 'https://raw.githubusercontent.com/golangci/golangci-lint/${GOCI_LINT_V}/install.sh' | sh -s -- -b ${GO_BIN} ${GOCI_LINT_V}
endif

.PHONY: format
format: ## Format source code with gofmt and golangci-lint
	gofmt -s -w .
	golangci-lint run --fix -v ./...

.PHONY: tidy
tidy:
	go mod tidy -v
