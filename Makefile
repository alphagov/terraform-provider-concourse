.PHONY: build
build: clean terraform-provider-concourse

.PHONY: clean
clean:
	go clean

terraform-provider-concourse:
	go build -o terraform-provider-concourse

.PHONY: install
install: terraform-provider-concourse
	@mkdir -p ~/.terraform.d/plugins/$$(uname | tr '[:upper:]' '[:lower:]' | tr -d '[:digit:]')_amd64
	@cp terraform-provider-concourse ~/.terraform.d/plugins/$$(uname | tr '[:upper:]' '[:lower:]' | tr -d '[:digit:]')_amd64
	@echo Installed terraform provider into ~/.terraform.d/plugins/$$(uname | tr '[:upper:]' '[:lower:]' | tr -d '[:digit:]')_amd64

.PHONY: integration-tests
integration-tests:
	@sh keys/generate
	go test -count 1 -v ./integration
