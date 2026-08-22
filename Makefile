.PHONY: build frontend backend run clean docker

# 构建前端
frontend:
	cd frontend && npm install && npm run build

# 构建后端
backend:
	go build -ldflags="-s -w" -o mayfly ./cmd/server/

# 构建全部
build: frontend backend

# 运行
run:
	./mayfly --config configs/config.yaml

# 清理
clean:
	rm -f mayfly mayfly.exe
	rm -rf frontend/dist web/dist

# Docker 构建
docker:
	docker build -t mayfly:latest .

# Docker Compose 启动
docker-up:
	docker compose up -d

# Docker Compose 停止
docker-down:
	docker compose down
