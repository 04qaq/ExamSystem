# exam-server — 纯 Go 二进制，配置见 config/config.yaml，可用环境变量覆盖（见 docker-compose）
FROM golang:1.25-bookworm AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/exam-server ./cmd/server

FROM debian:bookworm-slim

RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates tzdata \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=build /out/exam-server /app/exam-server
COPY config/config.yaml /app/config/config.yaml

EXPOSE 8080/tcp 43210/tcp 43210/udp

ENTRYPOINT ["/app/exam-server"]
