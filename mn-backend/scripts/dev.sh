#!/bin/bash

# 快速开发脚本
# 用法: ./scripts/dev.sh [命令] [环境]
# 命令:
#   build  - 编译项目
#   run    - 运行项目
#   dev    - 开发模式 (编译并运行)
#   clean  - 清理构建文件
#   hot    - 热重载模式
# 环境: dev (默认) | test | prod

set -e

PROJECT_NAME="moonick"

# Windows 下可执行文件需要 .exe 后缀
is_windows=false
if [[ "${OS:-}" == "Windows_NT" ]]; then
    is_windows=true
else
    case "$(uname -s 2>/dev/null || echo unknown)" in
        CYGWIN*|MINGW*|MSYS*)
            is_windows=true
            ;;
    esac
fi

if [[ "$is_windows" == "true" ]]; then
    BINARY_NAME="tmp/main.exe"
else
    BINARY_NAME="tmp/main"
fi

# 确保 tmp 目录存在
mkdir -p tmp

# 设置 Go 模块代理（中国镜像，更快）
export GOPROXY=https://goproxy.cn,direct
export GOSUMDB=sum.golang.google.cn

# 设置环境变量（默认为 dev）
export MOONICK_ENV="${2:-${MOONICK_ENV:-dev}}"

# 验证环境变量
if [[ ! "$MOONICK_ENV" =~ ^(dev|test|prod)$ ]]; then
    echo "❌ 错误: 无效的环境 '$MOONICK_ENV'，只允许: dev, test, prod"
    exit 1
fi

echo "🌍 当前环境: $MOONICK_ENV"

case "${1:-dev}" in
    "build")
        echo "🔨 编译项目..."
        time go build -o $BINARY_NAME ./cmd
        echo "✅ 编译完成: $BINARY_NAME"
        ;;
    "run")
        if [ ! -f "$BINARY_NAME" ]; then
            echo "⚠️  二进制文件不存在，正在编译..."
            ./scripts/dev.sh build
        fi
        echo "🚀 启动服务器..."
        ./$BINARY_NAME
        ;;
    "dev")
        echo "🔄 开发模式: 编译并运行..."
        ./scripts/dev.sh build
        ./scripts/dev.sh run
        ;;
    "clean")
        echo "🧹 清理构建文件..."
        rm -f $BINARY_NAME
        echo "✅ 清理完成"
        ;;
    "hot")
        echo "🔥 热重载模式 (需要安装 air)..."
        # 添加 GOPATH/bin 到 PATH
        export PATH="$(go env GOPATH)/bin:$PATH"
        if ! command -v air &> /dev/null; then
            echo "正在安装 air..."
            GOPROXY=https://goproxy.cn,direct go install github.com/air-verse/air@latest
        fi

        # 显式指定构建入口，避免在项目根目录构建导致 "no Go files"。
        AIR_BUILD_CMD="go build -o ./$BINARY_NAME ./cmd"
        AIR_BUILD_BIN="./$BINARY_NAME"
        echo "🛠️  air build cmd: $AIR_BUILD_CMD"
        echo "▶️  air run bin:   $AIR_BUILD_BIN"

        air --build.cmd "$AIR_BUILD_CMD" --build.bin "$AIR_BUILD_BIN"
        ;;
    *)
        echo "用法: $0 [build|run|dev|clean|hot] [环境]"
        echo ""
        echo "命令说明:"
        echo "  build  - 编译项目"
        echo "  run    - 运行已编译的项目"
        echo "  dev    - 开发模式 (编译并运行)"
        echo "  clean  - 清理构建文件"
        echo "  hot    - 热重载模式 (自动重启)"
        echo ""
        echo "环境参数 (可选):"
        echo "  dev    - 开发环境 (默认)"
        echo "  test   - 测试环境"
        echo "  prod   - 生产环境"
        echo ""
        echo "示例:"
        echo "  $0 hot          # 使用 dev 环境热重载"
        echo "  $0 hot prod     # 使用 prod 环境热重载"
        echo "  $0 build test   # 使用 test 环境编译"
        exit 1
        ;;
esac