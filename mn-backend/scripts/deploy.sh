#!/bin/bash

# 设置错误时退出
set -e

# 获取脚本所在目录的上一级目录（项目根目录）
PROJECT_ROOT=$(cd "$(dirname "$0")/.." && pwd)
cd "$PROJECT_ROOT"

echo "🚀 开始构建项目..."

# 1. 检查并创建 public 目录
if [ ! -d "public" ]; then
    echo "📂 创建 public 目录..."
    mkdir -p public
fi

# 2. 生成 meta.json
echo "📄 生成 public/meta.json..."
CURRENT_TIME=$(date '+%Y-%m-%d %H:%M:%S')
echo "{\"deployTime\": \"$CURRENT_TIME\"}" > public/meta.json

# 3. 编译项目
echo "🔨 正在编译..."
# 设置环境变量进行交叉编译（如需在本机运行可去掉这些变量，这里保留原有逻辑）
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

go build -o moonick ./cmd/main.go

if [ $? -eq 0 ]; then
    echo "✅ 构建成功！"
    echo "📍 输出文件: $PROJECT_ROOT/moonick"
    echo "📅 部署时间: $CURRENT_TIME"

    echo "📤 开始上传文件到服务器..."
    rsync -avz --progress --partial ./moonick qyhever:/opt/apps/moonick-backend
    rsync -avz --progress --partial ./public qyhever:/opt/apps/moonick-backend
    rsync -avz --progress --partial ./internal/config/prod.yml qyhever:/opt/apps/moonick-backend
    echo "✅ 上传完成！"
else
    echo "❌ 构建失败！"
    exit 1
fi
#!/bin/bash

# 设置错误时退出
set -e

# 获取脚本所在目录的上一级目录（项目根目录）
PROJECT_ROOT=$(cd "$(dirname "$0")/.." && pwd)
cd "$PROJECT_ROOT"

echo "🚀 开始构建项目..."

# 1. 检查并创建 public 目录
if [ ! -d "public" ]; then
    echo "📂 创建 public 目录..."
    mkdir -p public
fi

# 2. 生成 meta.json
echo "📄 生成 public/meta.json..."
CURRENT_TIME=$(date '+%Y-%m-%d %H:%M:%S')
echo "{\"deployTime\": \"$CURRENT_TIME\"}" > public/meta.json

# 3. 编译项目
echo "🔨 正在编译..."
# 设置环境变量进行交叉编译（如需在本机运行可去掉这些变量，这里保留原有逻辑）
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

go build -o moonick ./cmd/main.go

if [ $? -eq 0 ]; then
    echo "✅ 构建成功！"
    echo "📍 输出文件: $PROJECT_ROOT/moonick"
    echo "📅 部署时间: $CURRENT_TIME"

    echo "📤 开始上传文件到服务器..."
    rsync -avz --progress --partial ./moonick qyhever:/opt/apps/moonick-backend
    rsync -avz --progress --partial ./public qyhever:/opt/apps/moonick-backend
    rsync -avz --progress --partial ./internal/config/app.yml qyhever:/opt/apps/moonick-backend
    rsync -avz --progress --partial ./internal/config/prod.yml qyhever:/opt/apps/moonick-backend
    echo "✅ 上传完成！"
else
    echo "❌ 构建失败！"
    exit 1
fi
