#!/usr/bin/env bash

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DEPLOY_HOST="${DEPLOY_HOST:-qyhever}"
DEPLOY_PATH="${DEPLOY_PATH:-/usr/share/nginx/html/moonick}"
BUILD_DIR="${BUILD_DIR:-dist}"

log() {
  printf '[deploy] %s\n' "$*"
}

fail() {
  printf '[deploy] ERROR: %s\n' "$*" >&2
  exit 1
}

require_command() {
  local command_name="$1"
  command -v "$command_name" >/dev/null 2>&1 || fail "缺少依赖命令: ${command_name}"
}

log "项目根目录: ${PROJECT_ROOT}"
cd "${PROJECT_ROOT}"

require_command npm
require_command rsync
require_command ssh

[ -f package.json ] || fail "未找到 package.json，请确认脚本位于 mn-frontend-h5/scripts 目录下"

log "检查远端主机连通性: ${DEPLOY_HOST}"
ssh -o BatchMode=yes -o ConnectTimeout=10 "${DEPLOY_HOST}" "mkdir -p '${DEPLOY_PATH}'" \
  || fail "无法连接远端主机或创建目录: ${DEPLOY_HOST}:${DEPLOY_PATH}"

log "开始构建 H5"
npm run build

[ -d "${BUILD_DIR}" ] || fail "构建完成后未找到产物目录: ${PROJECT_ROOT}/${BUILD_DIR}"
[ -f "${BUILD_DIR}/index.html" ] || fail "构建产物不完整，缺少 ${BUILD_DIR}/index.html"

log "开始同步构建产物到 ${DEPLOY_HOST}:${DEPLOY_PATH}"
rsync -avz --delete "${BUILD_DIR}/" "${DEPLOY_HOST}:${DEPLOY_PATH}/"

log "部署完成: ${DEPLOY_HOST}:${DEPLOY_PATH}"
