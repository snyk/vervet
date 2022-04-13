GOMOD=go mod
GOBUILD=go build
GOTEST=go test

.PHONY: run fmt lint tidy test test-coverage clean test-localstack

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: clean
clean:
	rm -f server && go clean

.PHONY: lint
lint:
	golangci-lint run  --fix -v ./...

.PHONY: lint-docker
lint-docker:
	docker run --rm -v $(shell pwd):/vervet-underground -w /vervet-underground golangci/golangci-lint:v1.42.1 golangci-lint run -v ./...

tidy:
	$(GOMOD) tidy -v
test:
	go test ./... -count=1 -race

build:
	go build server.go
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
	go test ./... -count=1 -ginkgo.failFast -coverprofile=covfile
	go tool cover -html=covfile
	rm -f covfile

.PHONY: start
start:
	go run server.go

.PHONY: test-localstack
test-localstack:
	docker-compose up --no-start localstack
	docker-compose start localstack
	go test ./... -count=1 -race
	docker-compose down