#!/bin/bash

# 检查swag命令是否存在
if ! command -v swag &> /dev/null; then
    echo "未找到swag命令，正在安装..."
    go install github.com/swaggo/swag/cmd/swag@latest
    
    # 检查安装是否成功
    if ! command -v swag &> /dev/null; then
        echo "swag安装失败，可能是PATH环境变量问题"
        
        # 尝试直接使用GOPATH中的swag
        GOPATH=$(go env GOPATH)
        SWAG_PATH="$GOPATH/bin/swag"
        
        if [ -f "$SWAG_PATH" ]; then
            echo "在 $SWAG_PATH 找到swag，将直接使用此路径"
            
            # 清空现有的swag-docs目录
            rm -rf swag-docs
            
            # 使用完整路径执行swag
            "$SWAG_PATH" init -g cmd/app/main.go -o swag-docs
            
            echo "Swagger文档已生成到 swag-docs/ 目录"
            exit 0
        else
            echo "在 $GOPATH/bin 中未找到swag，请确保安装成功并将 $GOPATH/bin 添加到PATH环境变量"
            echo "可以尝试: export PATH=\$PATH:\$(go env GOPATH)/bin"
            exit 1
        fi
    fi
    echo "swag安装成功"
fi

# 清空现有的swag-docs目录
rm -rf swag-docs

# 生成Swagger文档
swag init -g cmd/app/main.go -o swag-docs

echo "Swagger文档已生成到 swag-docs/ 目录" 