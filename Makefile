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
	@go build -o build/signctrl *.go

# Build for linux
build-linux:
	@echo "--> Building SignCTRL for linux/amd64..."
	GOOS=linux GOARCH=amd64 $(MAKE) build

# Install the binary to $GOPATH/bin
install:
	@echo "--> Installing SignCTRL to "$(shell go env GOPATH)"/bin..."
	@go build -o $(shell go env GOPATH)/bin/signctrl *.go

# Download dependencies
go-mod-cache:
	@echo "--> Downloading dependencies for SignCTRL..."
	@go mod download

# Verify dependencies
go.sum:
	@echo "--> Ensuring dependencies for SignCTRL have not been modified..."
	@go mod verify