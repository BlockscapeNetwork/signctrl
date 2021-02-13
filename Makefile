# Build for local system
build:
	@echo "--> Building SignCTRL..."
	go build -o build/signctrl *.go

# Build for linux
build-linux:
	@echo "--> Building SignCTRL for linux/amd64..."
	GOOS=linux GOARCH=amd64 $(MAKE) build

# Install the binary to $GOPATH/bin
install:
	@echo "--> Installing SignCTRL..."
	@go build -o $(shell go env GOPATH)/bin/sc *.go

# Download dependencies
go-mod-cache:
	@echo "--> Downloading dependencies for SignCTRL..."
	@go mod download

# Verify dependencies
go.sum:
	@echo "--> Ensuring dependencies for SignCTRL have not been modified..."
	@go mod verify