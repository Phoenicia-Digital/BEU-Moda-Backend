PROJECT_NAME=Phoenicia-Digital-Base-API
MAIN_PACKAGE=main.go
BUILD_DIR=dist

.PHONY: all build clean

all: build

build:
	@echo "Building $(PROJECT_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -v -o $(BUILD_DIR) $(MAIN_PACKAGE)
	@mkdir -p $(BUILD_DIR)/config
	@cp config/.env $(BUILD_DIR)/config

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)

run:
	@./$(BUILD_DIR)/main