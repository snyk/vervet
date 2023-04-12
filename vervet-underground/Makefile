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
	go build tidy -v

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
	go test ./... -count=1 -coverprofile=covfile
	go tool cover -html=covfile
	rm -f covfile

.PHONY: start
start:
	go run server.go
