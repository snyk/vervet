.PHONY: build
build:
	go build -a -o vervet ./cmd/vervet

# Run go mody tidy yourself

#----------------------------------------------------------------------------------
# Check for updates to packages in remote OSS repositories and update go.mod AND
# go.sum to match changes. Then download the all the dependencies
# This catches when your app has colliding versions of packages during updates
#----------------------------------------------------------------------------------
.PHONY: update-deps
update-deps:
	go get -d -u ./...

# go mod download yourself if you don't need to update

#----------------------------------------------------------------------------------
#  Ignores the test cache and forces a full test suite execution
#----------------------------------------------------------------------------------
.PHONY: test
test:
	go test ./... -count=1

.PHONY: test-coverage
test-coverage:
	go test ./... -count=1 -coverprofile=covfile
	go tool cover -html=covfile
	rm -f covfile

.PHONY: clean
clean:
	$(RM) vervet