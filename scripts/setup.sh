#!/bin/bash

# 设置颜色
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 检查是否安装了Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Docker未安装，请先安装Docker${NC}"
    exit 1
fi

# 检查是否安装了Docker Compose
if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}Docker Compose未安装，请先安装Docker Compose${NC}"
    exit 1
fi

# 创建开发环境
echo -e "${GREEN}正在创建开发环境...${NC}"

# 创建docker-compose.yml文件
cat > docker-compose.yml << EOF
version: '3'

services:
  postgres:
    image: postgres:17-alpine
    container_name: go-rest-postgres
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: restapi
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s