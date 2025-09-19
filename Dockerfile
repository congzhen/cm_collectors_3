FROM alpine:3.14

# 安装必要的依赖
RUN apk add --no-cache ca-certificates

# 安装ffmpeg（包含ffmpeg和ffprobe）
RUN apk add --no-cache ffmpeg

# 创建工作目录
WORKDIR /app

# 复制预编译的服务器可执行文件
COPY build/cm_collectors_server .


# 设置执行权限
RUN chmod +x ./cm_collectors_server

# 暴露端口
EXPOSE 12345

# 设置环境变量
ENV GIN_MODE=release

# 启动应用
CMD ["./cm_collectors_server"]