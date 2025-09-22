@echo off
REM 自动注册cmcollectorsvideoplay协议

REM 设置UTF-8代码页以避免乱码
chcp 65001 >nul

REM 检查是否以管理员权限运行
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo 错误: 此脚本需要以管理员权限运行。
    pause
    exit /b
)


REM 获取当前目录
set CURRENT_DIR=%~dp0

REM 移除末尾的反斜杠
if "%CURRENT_DIR:~-1%"=="\" set CURRENT_DIR=%CURRENT_DIR:~0,-1%

echo 正在注册协议，当前目录: %CURRENT_DIR%

REM 创建临时注册表文件
set TEMP_REG=%TEMP%\cmcollectorsvideoplay.reg

echo Windows Registry Editor Version 5.00 > "%TEMP_REG%"
echo. >> "%TEMP_REG%"
echo [HKEY_CLASSES_ROOT\cmcollectorsvideoplay] >> "%TEMP_REG%"
echo @="URL:Video Player Protocol" >> "%TEMP_REG%"
echo "URL Protocol"="" >> "%TEMP_REG%"
echo. >> "%TEMP_REG%"
echo [HKEY_CLASSES_ROOT\cmcollectorsvideoplay\shell] >> "%TEMP_REG%"
echo. >> "%TEMP_REG%"
echo [HKEY_CLASSES_ROOT\cmcollectorsvideoplay\shell\open] >> "%TEMP_REG%"
echo. >> "%TEMP_REG%"
echo [HKEY_CLASSES_ROOT\cmcollectorsvideoplay\shell\open\command] >> "%TEMP_REG%"
echo @="%CURRENT_DIR:\=\\%\\cm_collectors_video_caller.exe \\\"%%1\\\"" >> "%TEMP_REG%"

echo 导入注册表项...
reg import "%TEMP_REG%"

echo 删除临时文件...
del "%TEMP_REG%"

echo 协议注册完成！
pause