#!/bin/bash
set -e

# 首先生成最新的Swagger文档
./scripts/swagger.sh

# 检查 air 是否已安装
if ! command -v air &> /dev/null; then
    echo "Installing air..."
    go install github.com/air-verse/air@latest
fi

# 获取 GOPATH
GOPATH=$(go env GOPATH)
AIR_BIN="$GOPATH/bin/air"

# 检查 air 是否成功安装
if [ ! -f "$AIR_BIN" ]; then
    echo "Error: air installation failed"
    exit 1
fi

# 使用完整路径运行 air
echo "Starting air..."
"$AIR_BIN" -c .air.toml 