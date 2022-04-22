GOMOD=go mod
GOBUILD=go build
GOTEST=go test

CLUSTER_NAME := vervet-underground

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
	docker run --rm -v $(shell pwd):/vervet-underground -w /vervet-underground golangci/golangci-lint:v1.43.0 golangci-lint run --fix -v ./...

.PHONY: tidy
tidy:
	$(GOMOD) tidy -v

.PHONY: test
test:
	go test ./... -count=1 -race

.PHONY: build
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
	go test ./... -count=1 -race || true
	docker-compose down

.PHONY: cluster/start
cluster/start:
	k3d cluster list ${CLUSTER_NAME} || k3d cluster create ${CLUSTER_NAME} --registry-create vu-registry

.PHONY: cluster/delete
cluster/delete:
	k3d cluster delete ${CLUSTER_NAME}

.PHONY: cluster/context
cluster/context:
	kubectl config use-context k3d-${CLUSTER_NAME}
