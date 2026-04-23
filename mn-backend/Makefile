# Makefile for moonick project

.PHONY: build run dev clean test install-air hot help

# 变量定义
BINARY_NAME=moonick
MAIN_PATH=./cmd
BUILD_DIR=./tmp

# 设置 Go 模块代理（中国镜像）
export GOPROXY=https://goproxy.cn,direct
export GOSUMDB=sum.golang.google.cn

# 默认目标
.DEFAULT_GOAL := help

## build: 编译项目
build:
	@echo "🔨 编译项目..."
	@time go build -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ 编译完成: $(BINARY_NAME)"

## run: 运行已编译的项目
run: build
	@echo "🚀 启动服务器..."
	@./$(BINARY_NAME)

## dev: 开发模式 (编译并运行)
dev: build
	@./$(BINARY_NAME)

## clean: 清理构建文件
clean:
	@echo "🧹 清理构建文件..."
	@rm -f $(BINARY_NAME)
	@rm -rf tmp/
	@echo "✅ 清理完成"

## test: 运行测试
test:
	@echo "🧪 运行测试..."
	@go test -v ./...

## install-air: 安装 air 热重载工具
install-air:
	@echo "📦 安装 air 热重载工具..."
	@GOPROXY=https://goproxy.cn,direct go install github.com/air-verse/air@latest
	@echo "✅ air 安装完成"

## hot: 热重载模式 (需要先安装 air)
hot:
	@echo "🔥 启动热重载模式..."
	@if ! command -v air >/dev/null 2>&1; then \
		if [ ! -f "`go env GOPATH`/bin/air" ]; then \
			echo "⚠️  air 未安装，正在安装..."; \
			$(MAKE) install-air; \
		fi; \
	fi
	@export PATH="`go env GOPATH`/bin:$$PATH" && \
	air \
		--build.cmd "go build -o ./tmp/$(BINARY_NAME) ./cmd/main.go" \
		--build.bin "./tmp/$(BINARY_NAME)" \
		--build.delay "300"

## deps: 整理依赖
deps:
	@echo "📦 整理依赖..."
	@go mod tidy
	@go mod download
	@echo "✅ 依赖整理完成"

## help: 显示帮助信息
help:
	@echo "moonick 项目 Makefile"
	@echo ""
	@echo "可用命令:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
	@echo ""
	@echo "快速开始:"
	@echo "  make dev     # 编译并运行项目"
	@echo "  make hot     # 热重载模式 (推荐开发使用)"
	@echo "  make test    # 运行测试"