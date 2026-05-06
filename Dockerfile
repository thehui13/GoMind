# --- 第一阶段：构建阶段 (Builder Stage) ---
    FROM golang:1.24-bookworm AS builder

    # 1. 安装下载和解压所需的工具
    RUN apt-get update && apt-get install -y wget tar
    
    WORKDIR /app
    
    # 2. 下载适用于 Apple Silicon (ARM64) 的 ONNX Runtime 动态库
    # 注意：作为 MacBook Air 用户，必须使用 aarch64 版本 [cite: 34]
    RUN wget https://github.com/microsoft/onnxruntime/releases/download/v1.20.1/onnxruntime-linux-aarch64-1.20.1.tgz \
        && tar -xzf onnxruntime-linux-aarch64-1.20.1.tgz \
        && cp onnxruntime-linux-aarch64-1.20.1/lib/libonnxruntime.so* /usr/local/lib/ \
        && ldconfig
    
    # 3. 复制依赖文件并下载 (利用 Docker 缓存) [cite: 5, 20]
    COPY go.mod go.sum ./
    RUN go mod download
    
    # 4. 复制完整的源代码 [cite: 5, 20]
    COPY . .
    
    # 5. 启用 CGO 进行编译
    # 必须设置 CGO_ENABLED=1，否则会报 "build constraints exclude all Go files" 错误 [cite: 3, 7, 191]
    RUN CGO_ENABLED=1 GOOS=linux go build -o /bin/gomind ./main.go
    
    
    # --- 第二阶段：运行阶段 (Runtime Stage) ---
    FROM debian:bookworm-slim
    
    WORKDIR /app
    
    # 6. 安装运行时必要的证书和库管理工具
    RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates \
        && rm -rf /var/lib/apt/lists/*
    
    # 7. 关键：从构建阶段将动态库复制到最终镜像中 [cite: 13, 18]
    # 如果没有这些 .so 文件，程序启动时会报 "no such file or directory" [cite: 14, 33]
    COPY --from=builder /usr/local/lib/libonnxruntime.so* /usr/local/lib/
    RUN ldconfig
    
    # 8. 复制编译好的二进制文件和配置文件
    COPY --from=builder /bin/gomind /app/gomind
    COPY --from=builder /app/config /app/config
    
    # 暴露后端端口
    EXPOSE 9090
    
    # 启动程序
    CMD ["/app/gomind"]