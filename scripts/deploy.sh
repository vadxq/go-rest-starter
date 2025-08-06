#!/bin/bash

# 生产环境部署脚本

set -e  # 出错即停止

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}开始部署...${NC}"

# 检查环境变量
check_env() {
    local vars=("DB_HOST" "DB_USER" "DB_PASSWORD" "DB_NAME" "JWT_SECRET")
    for var in "${vars[@]}"; do
        if [ -z "${!var}" ]; then
            echo -e "${RED}错误: 环境变量 $var 未设置${NC}"
            exit 1
        fi
    done
}

# 构建应用
build() {
    echo -e "${YELLOW}构建应用...${NC}"
    go build -ldflags="-s -w" -o bin/app cmd/app/main.go
    echo -e "${GREEN}构建完成${NC}"
}

# 运行数据库迁移
migrate() {
    echo -e "${YELLOW}运行数据库迁移...${NC}"
    # go run cmd/migrate/main.go up
    echo -e "${GREEN}迁移完成${NC}"
}

# 健康检查
health_check() {
    echo -e "${YELLOW}执行健康检查...${NC}"
    local max_attempts=30
    local attempt=0
    
    while [ $attempt -lt $max_attempts ]; do
        if curl -f http://localhost:${PORT:-8080}/health > /dev/null 2>&1; then
            echo -e "${GREEN}健康检查通过${NC}"
            return 0
        fi
        
        attempt=$((attempt + 1))
        echo "等待服务启动... ($attempt/$max_attempts)"
        sleep 2
    done
    
    echo -e "${RED}健康检查失败${NC}"
    return 1
}

# 优雅停止旧进程
stop_old() {
    if [ -f "app.pid" ]; then
        OLD_PID=$(cat app.pid)
        if kill -0 $OLD_PID 2>/dev/null; then
            echo -e "${YELLOW}停止旧进程 (PID: $OLD_PID)...${NC}"
            kill -TERM $OLD_PID
            sleep 5
        fi
    fi
}

# 启动新进程
start() {
    echo -e "${YELLOW}启动应用...${NC}"
    CONFIG_PATH=configs/config.production.yaml nohup ./bin/app > logs/app.out 2>&1 &
    echo $! > app.pid
    echo -e "${GREEN}应用已启动 (PID: $(cat app.pid))${NC}"
}

# 主流程
main() {
    check_env
    build
    # migrate
    stop_old
    start
    health_check
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}部署成功！${NC}"
    else
        echo -e "${RED}部署失败！${NC}"
        exit 1
    fi
}

main