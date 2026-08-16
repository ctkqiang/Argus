# syntax=docker/dockerfile:1.7
# Argus (观枢) 多阶段构建 Dockerfile
# 用法：
#   docker build --target gateway    -t argus-gateway:dev    .
#   docker build --target controller -t argus-controller:dev .
# 生产镜像使用 distroless 静态基础，无 shell 无包管理器，攻击面最小。

# ---------- builder ----------
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG LDFLAGS="-s -w -X main.version=${VERSION}"

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="${LDFLAGS}" -o /out/argus-gateway    ./cmd/argus-gateway

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="${LDFLAGS}" -o /out/argus-controller ./cmd/argus-controller

# ---------- gateway ----------
FROM gcr.io/distroless/static:nonroot AS gateway
COPY --from=builder /out/argus-gateway /argus-gateway
USER nonroot:nonroot
EXPOSE 8443 9090
ENTRYPOINT ["/argus-gateway"]

# ---------- controller ----------
FROM gcr.io/distroless/static:nonroot AS controller
COPY --from=builder /out/argus-controller /argus-controller
USER nonroot:nonroot
EXPOSE 8444 9090
ENTRYPOINT ["/argus-controller"]
