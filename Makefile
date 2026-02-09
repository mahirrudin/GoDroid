BINARY_NAME=godroid
CMD_PATH=./cmd/godroid
BUILD_DIR=bin

.PHONY: all build clean run linux darwin windows

all: build

build:
	@echo "Building for current OS..."
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)

clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)

run: build
	@echo "Running..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Cross compilation
linux:
	@echo "Building for Linux (amd64)..."
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/linux/$(BINARY_NAME) $(CMD_PATH)

darwin:
	@echo "Building for macOS (amd64 & arm64)..."
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/darwin/amd64/$(BINARY_NAME) $(CMD_PATH)
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/darwin/arm64/$(BINARY_NAME) $(CMD_PATH)

windows:
	@echo "Building for Windows (amd64)..."
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/windows/$(BINARY_NAME).exe $(CMD_PATH)

multi: linux darwin windows
	@echo "Multi-platform build complete."

test:
	@echo "Running tests..."
	go test -v ./...

cover:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

vet:
	@echo "Running go vet..."
	go vet ./...

lint: vet
	@echo "Linting complete."

