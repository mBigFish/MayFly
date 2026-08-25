# 构建阶段
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o mayfly .

# 运行阶段
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

COPY --from=builder /app/mayfly .
COPY --from=builder /app/config/config.yaml ./config/config.yaml
COPY --from=builder /app/web ./web
COPY --from=builder /app/payloads ./payloads

RUN mkdir -p data

EXPOSE 8080

CMD ["./mayfly"]
