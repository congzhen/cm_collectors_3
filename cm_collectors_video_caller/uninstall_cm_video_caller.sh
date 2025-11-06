#!/bin/bash
# uninstall_cm_video_caller.sh

# 检查是否以root权限运行
if [[ $EUID -eq 0 ]]; then
   echo "请不要以root权限运行此脚本"
   exit 1
fi

echo "正在卸载cmcollectorsvideoplay协议..."

# 删除desktop文件
DESKTOP_FILE="$HOME/.local/share/applications/cmcollectorsvideoplay.desktop"
if [ -f "$DESKTOP_FILE" ]; then
    rm "$DESKTOP_FILE"
    echo "已删除desktop文件"
else
    echo "未找到desktop文件"
fi

# 更新desktop数据库
if command -v update-desktop-database &> /dev/null; then
    update-desktop-database $HOME/.local/share/applications/
fi

echo "协议卸载完成！"