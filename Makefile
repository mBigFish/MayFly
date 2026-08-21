.PHONY: build run test vet frontend-install frontend-dev frontend-build clean

# 编译后端
build:
	go build -o bin/webshell-manager ./cmd/server

# 运行后端
run:
	go run ./cmd/server

# 运行后端单元测试
test:
	go test ./...

# 静态检查
vet:
	go vet ./...

# 安装前端依赖
frontend-install:
	cd frontend && npm install

# 前端开发模式
frontend-dev:
	cd frontend && npm run dev

# 前端构建
frontend-build:
	cd frontend && npm run build

# 清理构建产物
clean:
	rm -rf bin data logs frontend/dist
