#!/bin/bash

# 首先生成最新的Swagger文档
./scripts/swagger.sh

# 使用air运行应用程序（需要先安装air: go install github.com/cosmtrek/air@latest）
air -c .air.toml 