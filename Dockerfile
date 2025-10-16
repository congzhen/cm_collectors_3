FROM ubuntu:20.04

# 避免交互式配置
ENV DEBIAN_FRONTEND=noninteractive

# 设置时区和本地化环境变量
ENV TZ=Asia/Shanghai \
    LANG=zh_CN.UTF-8

# 更新包索引并安装必要的依赖
RUN apt-get update && apt-get upgrade -y && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    locales \
    language-pack-zh-hans \
    fonts-wqy-zenhei \
    fonts-wqy-microhei \
    wget \
    gnupg

# 生成本地化
RUN locale-gen zh_CN.UTF-8

# 安装ffmpeg
RUN apt-get update && \
    apt-get install -y --no-install-recommends ffmpeg

# 安装Google Chrome
RUN wget -q -O - https://dl.google.com/linux/linux_signing_key.pub | apt-key add - && \
    echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" > /etc/apt/sources.list.d/google-chrome.list && \
    apt-get update && \
    apt-get install -y --no-install-recommends \
    google-chrome-stable \
    xvfb

# 清理包缓存以减小镜像大小
RUN apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# 创建工作目录
WORKDIR /app

# 复制预编译的服务器可执行文件
COPY build/cm_collectors_server .

# 复制scraper目录
COPY cm_collectors_server/scraper ./scraper

# 设置执行权限
RUN chmod +x ./cm_collectors_server

# 暴露端口
EXPOSE 12345

# 设置环境变量
ENV GIN_MODE=release \
    CHROME_BIN=/usr/bin/google-chrome \
    PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=true

# 启动应用
CMD ["./cm_collectors_server"]