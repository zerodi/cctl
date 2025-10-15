.PHONY: all help clean build build-linux-amd64 build-darwin-arm64 build-darwin-amd64 build-windows-amd64

APP_NAME ?= cctl
OUTPUT_DIR ?= dist
GO_FILES := $(shell find . -name '*.go' -type f -not -path "./vendor/*")

all: build

help:
	@echo "Available targets:"
	@echo "  make build                 # Build for current platform"
	@echo "  make build-linux-amd64     # Cross-compile for Linux amd64"
	@echo "  make build-darwin-amd64    # Cross-compile for macOS amd64"
	@echo "  make build-darwin-arm64    # Cross-compile for macOS arm64"
	@echo "  make build-windows-amd64   # Cross-compile for Windows amd64"
	@echo "  make clean                 # Remove build artifacts"

$(OUTPUT_DIR):
	mkdir -p $(OUTPUT_DIR)

build: $(OUTPUT_DIR) $(GO_FILES)
	GO111MODULE=on go build -o $(OUTPUT_DIR)/$(APP_NAME) .

build-linux-amd64: $(OUTPUT_DIR) $(GO_FILES)
	GO111MODULE=on GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(OUTPUT_DIR)/$(APP_NAME)-linux-amd64 .

build-darwin-amd64: $(OUTPUT_DIR) $(GO_FILES)
	GO111MODULE=on GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o $(OUTPUT_DIR)/$(APP_NAME)-darwin-amd64 .

build-darwin-arm64: $(OUTPUT_DIR) $(GO_FILES)
	GO111MODULE=on GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o $(OUTPUT_DIR)/$(APP_NAME)-darwin-arm64 .

build-windows-amd64: $(OUTPUT_DIR) $(GO_FILES)
	GO111MODULE=on GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o $(OUTPUT_DIR)/$(APP_NAME)-windows-amd64.exe .

clean:
	rm -rf $(OUTPUT_DIR)
