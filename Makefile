# FuckLang Makefile

APP_NAME = fucklang
CMD_PATH = ./cmd/fucklang
BUILD_DIR = build

# 默认目标
.PHONY: all build test clean release

all: build

# 本地编译（测试用）
build:
	go build -o $(APP_NAME) $(CMD_PATH)

# 运行测试
test:
	go clean -testcache
	go test -v ./...

# 交叉编译全平台二进制
release: clean
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	
	# Linux (amd64, arm64, riscv64)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(CMD_PATH)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 $(CMD_PATH)
	CGO_ENABLED=0 GOOS=linux GOARCH=riscv64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-linux-riscv64 $(CMD_PATH)
	
	# Windows (amd64, arm64)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe $(CMD_PATH)
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-windows-arm64.exe $(CMD_PATH)
	
	# macOS (Intel & Apple Silicon)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 $(CMD_PATH)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 $(CMD_PATH)
	
	@echo "Done! Binaries are in the $(BUILD_DIR)/ directory."

# 清理编译产物
clean:
	@rm -rf $(BUILD_DIR)
	@rm -f $(APP_NAME) $(APP_NAME).exe