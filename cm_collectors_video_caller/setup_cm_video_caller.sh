#!/bin/bash
# setup_cm_video_caller.sh

# 检查是否以root权限运行
if [[ $EUID -eq 0 ]]; then
   echo "请不要以root权限运行此脚本"
   exit 1
fi

# 获取当前目录
CURRENT_DIR=$(pwd)
EXECUTABLE="cm_collectors_video_caller"

echo "正在注册cmcollectorsvideoplay协议..."
echo "当前目录: $CURRENT_DIR"

# 创建.desktop文件
DESKTOP_FILE="$HOME/.local/share/applications/cmcollectorsvideoplay.desktop"

# 确保目录存在
mkdir -p "$HOME/.local/share/applications/"

# 写入desktop文件内容
cat > "$DESKTOP_FILE" << EOF
[Desktop Entry]
Name=CM Collectors Video Caller
Exec=$CURRENT_DIR/$EXECUTABLE %u
Type=Application
Terminal=false
StartupNotify=true
MimeType=x-scheme-handler/cmcollectorsvideoplay;
Icon=video-player
EOF

# 注册协议处理程序
xdg-mime default cmcollectorsvideoplay.desktop x-scheme-handler/cmcollectorsvideoplay

# 更新desktop数据库
if command -v update-desktop-database &> /dev/null; then
    update-desktop-database $HOME/.local/share/applications/
fi

echo "协议注册完成！"
echo "请确保$EXECUTABLE文件在当前目录中且具有执行权限"