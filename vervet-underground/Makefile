APP:=vervet-underground
GO_BIN=$(shell pwd)/.bin/go

SHELL:=env PATH=$(GO_BIN):$(PATH) $(SHELL)

GOCI_LINT_V?=v1.54.2

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: clean
clean:
	rm -f api scraper && go clean

.PHONY: lint
lint:
	golangci-lint run  --fix -v ./...

.PHONY: lint-docker
lint-docker:
	docker run --rm -v $(shell pwd):/vervet-underground -w /vervet-underground golangci/golangci-lint:${GOCI_LINT_V} golangci-lint run --fix -v ./...

.PHONY: build-docker
build-docker:
	docker build \
		-t ${APP}:${CIRCLE_WORKFLOW_ID} \
		-t gcr.io/snyk-main/${APP}:${CIRCLE_SHA1} .

.PHONY: tidy
tidy:
	go mod tidy -v

.PHONY: test
test:
	go test ./... -count=1 -race

.PHONY: build
build: api scraper

api:
	go build -o api cmd/api/main.go

scraper:
	go build -o scraper cmd/scraper/main.go
#----------------------------------------------------------------------------------
# Check for updates to packages in remote OSS repositories and update go.mod AND
# go.sum to match changes. Then download the all the dependencies
# This catches when your app has colliding versions of packages during updates
#----------------------------------------------------------------------------------
.PHONY: update-deps
update-deps:
	go get -d -u ./...

# go mod download yourself if you don't need to update

.PHONY: test-coverage
test-coverage:
	go test ./... -count=1 -coverprofile=covfile
	go tool cover -html=covfile
	rm -f covfile

.PHONY: start
start:
	go run cmd/scraper/main.go
	go run cmd/api/main.go

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
