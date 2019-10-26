.PHONY: build
build:
	go build -o terraform-provider-concourse

.PHONY: integration-tests
integration-tests:
	go test -count 1 -v ./integration
