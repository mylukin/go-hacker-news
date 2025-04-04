# Hacker News Podcast Generator Makefile

# 变量定义
APP_NAME = hnpodcast
BUILD_DIR = ./build
OUTPUT_DIR = ./output
MAIN_PATH = ./cmd/hnpodcast
GO_FILES = $(shell find . -name "*.go" -type f)

# 默认目标
.PHONY: all
all: clean build

# 创建必要的目录
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

$(OUTPUT_DIR):
	mkdir -p $(OUTPUT_DIR)

# 编译程序
.PHONY: build
build: $(BUILD_DIR)
	@echo "编译 $(APP_NAME)..."
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "完成! 可执行文件位于: $(BUILD_DIR)/$(APP_NAME)"

# 安装程序到GOPATH/bin
.PHONY: install
install:
	@echo "安装 $(APP_NAME)..."
	go install $(MAIN_PATH)
	@echo "安装完成!"

# 运行程序 (获取今天的故事)
.PHONY: run
run: $(OUTPUT_DIR)
	@echo "运行 $(APP_NAME)..."
	@if [ -f $(BUILD_DIR)/$(APP_NAME) ]; then \
		$(BUILD_DIR)/$(APP_NAME) -output $(OUTPUT_DIR); \
	else \
		echo "先构建程序..."; \
		$(MAKE) build; \
		$(BUILD_DIR)/$(APP_NAME) -output $(OUTPUT_DIR); \
	fi

# 以调试模式运行
.PHONY: debug
debug: $(OUTPUT_DIR)
	@echo "以调试模式运行 $(APP_NAME)..."
	@if [ -f $(BUILD_DIR)/$(APP_NAME) ]; then \
		$(BUILD_DIR)/$(APP_NAME) -output $(OUTPUT_DIR) -debug; \
	else \
		echo "先构建程序..."; \
		$(MAKE) build; \
		$(BUILD_DIR)/$(APP_NAME) -output $(OUTPUT_DIR) -debug; \
	fi

# 获取指定日期的故事
.PHONY: run-date
run-date: $(OUTPUT_DIR)
	@echo "请输入日期 (格式: YYYY-MM-DD):"
	@read date; \
	if [ -f $(BUILD_DIR)/$(APP_NAME) ]; then \
		$(BUILD_DIR)/$(APP_NAME) -output $(OUTPUT_DIR) -date $$date; \
	else \
		echo "先构建程序..."; \
		$(MAKE) build; \
		$(BUILD_DIR)/$(APP_NAME) -output $(OUTPUT_DIR) -date $$date; \
	fi

# 开发模式 (获取较少故事，提高速度)
.PHONY: dev
dev: $(OUTPUT_DIR)
	@echo "以开发模式运行 $(APP_NAME)..."
	@if [ -f $(BUILD_DIR)/$(APP_NAME) ]; then \
		$(BUILD_DIR)/$(APP_NAME) -output $(OUTPUT_DIR) -dev -max-stories 3; \
	else \
		echo "先构建程序..."; \
		$(MAKE) build; \
		$(BUILD_DIR)/$(APP_NAME) -output $(OUTPUT_DIR) -dev -max-stories 3; \
	fi

# 运行测试
.PHONY: test
test:
	@echo "运行测试..."
	go test -v ./...

# 清理构建文件
.PHONY: clean
clean:
	@echo "清理构建文件..."
	rm -rf $(BUILD_DIR)
	@echo "清理完成!"

# 清理输出文件
.PHONY: clean-output
clean-output:
	@echo "清理输出文件..."
	rm -rf $(OUTPUT_DIR)
	@echo "清理完成!"

# 更新依赖
.PHONY: deps
deps:
	@echo "更新依赖..."
	go mod tidy
	go mod download
	@echo "依赖更新完成!"

# 显示帮助信息
.PHONY: help
help:
	@echo "Hacker News 播客生成器 - 可用命令:"
	@echo "  make build       - 编译程序"
	@echo "  make install     - 安装程序到GOPATH/bin"
	@echo "  make run         - 运行程序 (获取今天的故事)"
	@echo "  make run-date    - 运行程序 (手动输入日期)"
	@echo "  make debug       - 以调试模式运行"
	@echo "  make dev         - 以开发模式运行 (只获取3个故事)"
	@echo "  make test        - 运行测试"
	@echo "  make clean       - 清理构建文件"
	@echo "  make clean-output - 清理输出文件"
	@echo "  make deps        - 更新依赖"
	@echo "  make help        - 显示此帮助信息" 