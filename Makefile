
OS = $$(go env GOOS)
ARCH = $$(go env GOARCH)

GENERATE_KEY := \
		docker run --rm -v $$PWD/keys:/keys --user $$(id -u):$$(id -g) \
		concourse/concourse:$${CONCOURSE_VERSION:-6.5.1} \
		generate-key

# shouldn't use CGO on binaries produced for the terraform registry
export CGO_ENABLED=0

.PHONY: build
build: clean terraform-provider-concourse

.PHONY: clean
clean:
	go clean

terraform-provider-concourse:
	go build -o terraform-provider-concourse

.PHONY: install
install: terraform-provider-concourse
	@mkdir -p ~/.terraform.d/plugins/$(OS)_$(ARCH)
	@cp terraform-provider-concourse ~/.terraform.d/plugins/$(OS)_$(ARCH)
	@echo Installed terraform provider into ~/.terraform.d/plugins/$(OS)_$(ARCH)

keys/web/session_signing_key:
	mkdir -p keys/web
	$(GENERATE_KEY) -t rsa -f /$@

keys/web/tsa_host_key:
	mkdir -p keys/web
	$(GENERATE_KEY) -t ssh -f/$@

keys/worker/worker_key:
	mkdir -p keys/worker
	$(GENERATE_KEY) -t ssh -f /$@

keys/worker/tsa_host_key.pub: keys/web/tsa_host_key
	mkdir -p keys/worker
	cp keys/web/tsa_host_key.pub $@

keys/web/authorized_worker_keys: keys/worker/worker_key
	mkdir -p keys/web
	cp keys/worker/worker_key.pub $@

# separate from `integration-tests` so it can be run as root without
# running the whole integration tests as root
.PHONY: integration-tests-prep-keys
integration-tests-prep-keys: keys/web/session_signing_key keys/web/tsa_host_key keys/worker/worker_key keys/worker/tsa_host_key.pub keys/web/authorized_worker_keys

.PHONY: integration-tests
integration-tests: integration-tests-prep-keys
	go test -count 1 -v ./integration

.PHONY: unit-tests
unit-tests:
	go test -count 1 -v ./pkg/provider
