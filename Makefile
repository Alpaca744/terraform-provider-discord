GO        ?= go
BINARY    := terraform-provider-discord
VERSION   ?= dev

.PHONY: build
build:
	$(GO) build -ldflags "-X main.version=$(VERSION)" -o $(BINARY) .

.PHONY: test
test:
	$(GO) test ./...

.PHONY: testrace
testrace:
	$(GO) test -race ./...

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: fmt
fmt:
	gofmt -w .

.PHONY: fmtcheck
fmtcheck:
	@test -z "$$(gofmt -l . | tee /dev/stderr)" || (echo "files need gofmt" && exit 1)

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: vuln
vuln:
	govulncheck ./...

# Generate Terraform Registry documentation from schemas, templates, and examples.
.PHONY: docs
docs:
	$(GO) run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate \
		--provider-name discord

# Acceptance tests require Discord credentials and TF_ACC=1.
# Usage: make testacc TESTS=./internal/service/role
TESTS ?= ./...
.PHONY: testacc
testacc:
	TF_ACC=1 $(GO) test $(TESTS) -v -timeout 120m

.PHONY: check
check: fmtcheck vet test
