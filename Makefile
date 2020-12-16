# Build for local system
build:
	@echo "--> Building pairmint..."
	go build -o build/pairmint *.go

# Build for linux
build-linux:
	@echo "--> Building pairmint for linux/amd64..."
	GOOS=linux GOARCH=amd64 $(MAKE) build

# Install the binary to $GOPATH/bin
install:
	@echo "--> Installing pairmint..."
	@go build -o $(shell go env GOPATH)/bin/pairmint *.go

# Download dependencies
deps:
	@echo "--> Downloading dependencies for pairmint..."
	@go mod download

# Verify dependencies
deps-verify:
	@echo "--> Ensuring dependencies for pairmint have not been modified..."
	@go mod verify