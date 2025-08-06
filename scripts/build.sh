#!/bin/bash
set -e

# 获取版本信息
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date +%Y-%m-%d_%H:%M:%S)
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建参数
LDFLAGS="-s -w"
OUTPUT_DIR="build"

# 创建输出目录
mkdir -p ${OUTPUT_DIR}

echo "Building Go REST Starter..."
echo "Version: ${VERSION}"
echo "Build Time: ${BUILD_TIME}"
echo "Commit: ${COMMIT}"

# 生成最新的Swagger文档
./scripts/swagger.sh

# 构建不同平台的二进制文件
echo "Building for Linux AMD64..."
GOOS=linux GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o ${OUTPUT_DIR}/app-linux-amd64 cmd/app/main.go

echo "Building for macOS AMD64..."
GOOS=darwin GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o ${OUTPUT_DIR}/app-darwin-amd64 cmd/app/main.go

echo "Building for macOS ARM64..."
GOOS=darwin GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o ${OUTPUT_DIR}/app-darwin-arm64 cmd/app/main.go

echo "Building for Windows AMD64..."
GOOS=windows GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o ${OUTPUT_DIR}/app-windows-amd64.exe cmd/app/main.go

# 构建当前平台的可执行文件
echo "Building for current platform..."
go build -ldflags="${LDFLAGS}" -o ${OUTPUT_DIR}/app cmd/app/main.go

echo "Build completed! Files are in ${OUTPUT_DIR}/"
ls -la ${OUTPUT_DIR}/