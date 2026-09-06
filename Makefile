APP_NAME := gito
SHORT_NAME := gt
VERSION := 0.0.5
DIST_DIR := dist
SRC_DIR := ./cmd/gito
GT_SRC_DIR := ./cmd/gt
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: all build install test clean cross-compile

all: build

build:
	@echo "Building $(APP_NAME) and $(SHORT_NAME)..."
	go build -ldflags="$(LDFLAGS)" -o $(APP_NAME) $(SRC_DIR)
	go build -ldflags="$(LDFLAGS)" -o $(SHORT_NAME) $(GT_SRC_DIR)

install:
	@echo "Installing $(APP_NAME) and $(SHORT_NAME) to /usr/local/bin..."
	@mkdir -p /usr/local/bin
	@go build -ldflags="$(LDFLAGS)" -o /usr/local/bin/$(APP_NAME) $(SRC_DIR)
	@ln -sf /usr/local/bin/$(APP_NAME) /usr/local/bin/$(SHORT_NAME)
	@chmod +x /usr/local/bin/$(APP_NAME)
	@echo "Installed successfully! You can now run '$(SHORT_NAME)' or '$(APP_NAME)'"

test:
	@echo "Running test suite..."
	go test -count=1 -v ./...

clean:
	@echo "Cleaning artifacts..."
	@rm -f $(APP_NAME) $(SHORT_NAME)
	@rm -rf $(DIST_DIR)

cross-compile: clean
	@echo "Cross-compiling $(APP_NAME) and $(SHORT_NAME) for Linux, macOS, and Windows..."
	@mkdir -p $(DIST_DIR)

	# macOS (Apple Silicon & Intel)
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME)-darwin-arm64 $(SRC_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME)-darwin-amd64 $(SRC_DIR)
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(SHORT_NAME)-darwin-arm64 $(GT_SRC_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(SHORT_NAME)-darwin-amd64 $(GT_SRC_DIR)

	# Linux (x86_64 & ARM64)
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME)-linux-amd64 $(SRC_DIR)
	GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME)-linux-arm64 $(SRC_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(SHORT_NAME)-linux-amd64 $(GT_SRC_DIR)
	GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(SHORT_NAME)-linux-arm64 $(GT_SRC_DIR)

	# Windows (x64 & ARM64)
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe $(SRC_DIR)
	GOOS=windows GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME)-windows-arm64.exe $(SRC_DIR)
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(SHORT_NAME)-windows-amd64.exe $(GT_SRC_DIR)
	GOOS=windows GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(SHORT_NAME)-windows-arm64.exe $(GT_SRC_DIR)

	@echo "Cross-compilation complete! Binaries in $(DIST_DIR)/:"
	@ls -la $(DIST_DIR)
