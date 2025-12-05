.PHONY: all build build-web build-go clean run test docker-build help

# 默认目标
all: build

# 帮助信息
help:
	@echo "Task Scheduler - 构建命令"
	@echo ""
	@echo "使用方法:"
	@echo "  make build        - 构建完整项目（Web + Go）"
	@echo "  make build-web    - 仅构建 Web UI"
	@echo "  make build-go     - 仅构建 Go 服务"
	@echo "  make run          - 运行服务"
	@echo "  make test         - 运行测试"
	@echo "  make clean        - 清理构建文件"
	@echo "  make docker-build - 构建 Docker 镜像"
	@echo ""

# 构建完整项目
build: build-web build-go
	@echo "✓ 构建完成！"

# 构建 Web UI
build-web:
	@echo "构建 Web UI..."
ifeq ($(OS),Windows_NT)
	@cmd /c scripts\build-web.bat
else
	@chmod +x scripts/build-web.sh
	@./scripts/build-web.sh
endif

# 构建 Go 服务
build-go:
	@echo "构建 Go 服务..."
	@go build -o task-scheduler.exe ./cmd/scheduler

# 清理构建文件
clean:
	@echo "清理构建文件..."
	@rm -rf web-dist
	@rm -f task-scheduler task-scheduler.exe
	@cd web && rm -rf node_modules dist
	@echo "✓ 清理完成！"

# 运行服务
run: build
	@echo "启动服务..."
	@./task-scheduler.exe -config config.local.yaml

# 运行测试
test:
	@echo "运行测试..."
	@go test -v ./...

# 构建 Docker 镜像
docker-build:
	@echo "构建 Docker 镜像..."
	@docker build -t task-scheduler:latest .

# 运行 Docker Compose
docker-up:
	@echo "启动 Docker Compose..."
	@docker-compose up -d

# 停止 Docker Compose
docker-down:
	@echo "停止 Docker Compose..."
	@docker-compose down

# 查看 Docker 日志
docker-logs:
	@docker-compose logs -f task-scheduler-1

# 开发模式 - 启动后端
dev-backend:
	@echo "启动后端开发服务器..."
	@go run ./cmd/scheduler -config config.local.yaml

# 开发模式 - 启动前端
dev-frontend:
	@echo "启动前端开发服务器..."
	@cd web && npm run dev

# 安装依赖
install-deps:
	@echo "安装依赖..."
	@go mod download
	@cd web && npm install

# 格式化代码
fmt:
	@echo "格式化代码..."
	@go fmt ./...
	@cd web && npm run format || true

# 代码检查
lint:
	@echo "代码检查..."
	@golangci-lint run || true
	@cd web && npm run lint || true
