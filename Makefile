APP_VERSION ?= $(shell git describe --abbrev=5 --dirty --tags --always)
GIT_COMMIT := $(shell git rev-parse --short=8 HEAD)
BUILD_TIME := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

BINDIR := $(PWD)/bin
OUTPUT_DIR := $(PWD)/_output

GOOS ?= $(shell uname -s | tr '[:upper:]' '[:lower:]')
GOARCH ?= amd64

LDFLAGS := $(LDFLAGS) -X github.com/bougou/go-ipmi/cmd/goipmi/commands.Version=$(APP_VERSION)
LDFLAGS := $(LDFLAGS) -X github.com/bougou/go-ipmi/cmd/goipmi/commands.Commit=$(GIT_COMMIT)
LDFLAGS := $(LDFLAGS) -X github.com/bougou/go-ipmi/cmd/goipmi/commands.BuildAt=$(BUILD_TIME)

PATH := $(BINDIR):$(PATH)
SHELL := env PATH='$(PATH)' /bin/sh

all: build

# Run tests
test: fmt vet
	@# Disable --race until https://github.com/kubernetes-sigs/controller-runtime/issues/1171 is fixed.
	ginkgo --randomizeAllSpecs --randomizeSuites --failOnPending --flakeAttempts=2 \
			--cover --coverprofile cover.out --trace --progress  $(TEST_ARGS)\
			./pkg/... ./cmd/...

# Build goipmi and goipmi-server binaries
build: fmt vet
	go build -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/goipmi ./cmd/goipmi
	go build -o $(OUTPUT_DIR)/goipmi-server ./cmd/goipmi-server

# Cross compiler
build-all: fmt vet
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -a -o $(OUTPUT_DIR)/goipmi-$(APP_VERSION)-linux-amd64 ./cmd/goipmi
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -a -o $(OUTPUT_DIR)/goipmi-$(APP_VERSION)-linux-arm64 ./cmd/goipmi
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -a -o $(OUTPUT_DIR)/goipmi-$(APP_VERSION)-darwin-amd64 ./cmd/goipmi
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -a -o $(OUTPUT_DIR)/goipmi-$(APP_VERSION)-darwin-arm64 ./cmd/goipmi

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

lint:
	$(BINDIR)/golangci-lint run --timeout 2m0s ./...

dependencies:
	test -d $(BINDIR) || mkdir $(BINDIR)
	GOBIN=$(BINDIR) go install github.com/onsi/ginkgo/ginkgo@v1.16.4

	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | bash -s -- -b $(BINDIR) latest

# ---------------------------------------------------------------------------
# E2E tests
# ---------------------------------------------------------------------------
#   make test-e2e-client  — goipmi → ipmi-simulator
#   make test-e2e-server  — ipmitool → goipmi-server
#   make test-e2e-self    — goipmi → goipmi-server
#   make test-e2e         — run all three

test-e2e-client: build
	./test/e2e/client_test.sh

test-e2e-server: build
	./test/e2e/server_test.sh

test-e2e-self: build
	./test/e2e/self_test.sh

test-e2e-chassis-codec: build
	./test/e2e/chassis_codec_test.sh

test-e2e-cipher: build
	./test/e2e/cipher_suite_test.sh

test-e2e: test-e2e-client test-e2e-server test-e2e-self test-e2e-chassis-codec test-e2e-cipher
