GOPATH := $(shell go env GOPATH)

# If building a release, checkout the version tag to get the correct version setting
ifneq ($(shell git symbolic-ref -q --short HEAD),)
	VERSION := unreleased-$(shell git symbolic-ref -q --short HEAD)-$(shell git rev-parse HEAD)
else
	VERSION := $(shell git describe --tags)
endif

GIT_COMMIT := $(shell git rev-list -1 HEAD)
LDFLAGS := -X github.com/BlockscapeNetwork/signctrl/cmd.SemVer=$(VERSION) \
	-X github.com/BlockscapeNetwork/signctrl/cmd.GitCommit=$(GIT_COMMIT)

# Allow users to pass additional flags via the conventional LDFLAGS variable
LDFLAGS += $(LDFLAGS)

# Build for local system
build:
	@echo "--> Building SignCTRL..."
	@go build -ldflags "$(LDFLAGS)" -o build/signctrl *.go
.PHONY: build

# Build for linux
build-linux:
	@echo "--> Building SignCTRL for linux/amd64..."
	GOOS=linux GOARCH=amd64 $(MAKE) build
.PHONY: build-linux

# Install the binary to $GOPATH/bin
install:
	@echo "--> Installing SignCTRL to "$(GOPATH)"/bin..."
	@go build -ldflags "$(LDFLAGS)" -o $(GOPATH)/bin/signctrl *.go
.PHONY: install

# Download dependencies
go-mod-cache: go.sum
	@echo "--> Downloading dependencies for SignCTRL..."
	@go mod download
.PHONY: go-mod-cache

# Verify dependencies
go.sum: go.mod
	@echo "--> Ensuring dependencies for SignCTRL have not been modified..."
	@go mod verify
.PHONY: go.sum