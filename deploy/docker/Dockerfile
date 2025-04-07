FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/app

FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app/server .
COPY configs/config.yaml ./configs/

EXPOSE 7001

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:7001/health || exit 1

CMD ["/app/server"]