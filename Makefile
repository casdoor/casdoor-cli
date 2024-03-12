APP_NAME := casdoor
APP_VERSION ?= 0.0.1
BUILD_DIR := bin
SRC_DIR := casdoor-cli
GO_FILES := $(wildcard $(SRC_DIR)/*.go)

.PHONY: build
build: $(GO_FILES)
	@echo "Building $(APP_NAME) version $(APP_VERSION)"
	@mkdir -p $(BUILD_DIR)
	@cd $(SRC_DIR) && go build -mod=mod -o ../$(BUILD_DIR)/$(APP_NAME) -ldflags "-X gitlab.com/sdv9972401/casdoor-cli-go/cmd.Version=$(APP_VERSION)"

.PHONY: clean
clean:
	@echo "Cleaning up build directory..."
	@rm -rf $(BUILD_DIR)