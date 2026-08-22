# ============================================================
# Stage 1: Build frontend
# ============================================================
FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci --silent
COPY frontend/ ./
RUN npm run build

# ============================================================
# Stage 2: Build backend
# ============================================================
FROM golang:1.23-alpine AS backend-builder

RUN apk add --no-cache git gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=frontend-builder /app/frontend/dist ./web/dist

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /mayfly ./cmd/server/

# ============================================================
# Stage 3: Final image
# ============================================================
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=backend-builder /mayfly .
COPY --from=backend-builder /app/payloads ./payloads
COPY --from=backend-builder /app/configs/config.yaml ./configs/config.yaml

ENV TZ=Asia/Shanghai
ENV GIN_MODE=release

EXPOSE 8080

VOLUME ["/app/data", "/app/logs"]

ENTRYPOINT ["/app/mayfly"]
CMD ["--config", "configs/config.yaml"]
