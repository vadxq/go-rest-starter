#!/bin/bash
set -e

# 默认配置
IMAGE_NAME="go-rest-starter"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "latest")
DOCKERFILE="deploy/docker/Dockerfile"

# 解析命令行参数
while [[ $# -gt 0 ]]; do
  case $1 in
    -n|--name)
      IMAGE_NAME="$2"
      shift 2
      ;;
    -v|--version)
      VERSION="$2"
      shift 2
      ;;
    -f|--file)
      DOCKERFILE="$2"
      shift 2
      ;;
    -h|--help)
      echo "Usage: $0 [OPTIONS]"
      echo "Options:"
      echo "  -n, --name IMAGE_NAME    Docker image name (default: go-rest-starter)"
      echo "  -v, --version VERSION    Image version tag (default: git describe)"
      echo "  -f, --file DOCKERFILE    Dockerfile path (default: deploy/docker/Dockerfile)"
      echo "  -h, --help              Show this help message"
      exit 0
      ;;
    *)
      echo "Unknown option $1"
      exit 1
      ;;
  esac
done

echo "Building Docker image..."
echo "Image: ${IMAGE_NAME}:${VERSION}"
echo "Dockerfile: ${DOCKERFILE}"

# 生成最新的Swagger文档
echo "Generating Swagger documentation..."
./scripts/swagger.sh

# 构建Docker镜像
echo "Building Docker image..."
docker build -t ${IMAGE_NAME}:${VERSION} -f ${DOCKERFILE} .

# 同时打上 latest 标签
if [[ "${VERSION}" != "latest" ]]; then
    docker tag ${IMAGE_NAME}:${VERSION} ${IMAGE_NAME}:latest
    echo "Tagged as ${IMAGE_NAME}:latest"
fi

echo "Docker build completed!"
echo "Images built:"
docker images ${IMAGE_NAME}

echo ""
echo "To run the container:"
echo "docker run -d -p 7001:7001 --name go-rest-starter-container \\"
echo "  -e APP_DATABASE_HOST=host.docker.internal \\"
echo "  -e APP_REDIS_HOST=host.docker.internal \\"
echo "  ${IMAGE_NAME}:${VERSION}"