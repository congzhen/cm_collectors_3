@echo off
REM 卸载cmcollectorsvideoplay协议注册

REM 设置UTF-8代码页以避免乱码
chcp 65001 >nul

REM 检查是否以管理员权限运行
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo 错误: 此脚本需要以管理员权限运行。
    pause
    exit /b
)



echo 正在卸载cmcollectorsvideoplay协议...

REM 方法1: 使用reg delete命令直接删除注册表项
echo 尝试使用reg delete命令删除注册表项...
reg delete "HKEY_CLASSES_ROOT\cmcollectorsvideoplay" /f
if %errorLevel% equ 0 (
    echo 成功删除注册表项!
) else (
    echo 警告: 使用reg delete命令删除注册表项失败，将尝试其他方法...
)

REM 方法2: 创建临时注册表文件用于删除项
set TEMP_REG=%TEMP%\uninstall_cmcollectorsvideoplay.reg

echo Windows Registry Editor Version 5.00 > "%TEMP_REG%"
echo. >> "%TEMP_REG%"
echo [-HKEY_CLASSES_ROOT\cmcollectorsvideoplay\shell\open\command] >> "%TEMP_REG%"
echo. >> "%TEMP_REG%"
echo [-HKEY_CLASSES_ROOT\cmcollectorsvideoplay\shell\open] >> "%TEMP_REG%"
echo. >> "%TEMP_REG%"
echo [-HKEY_CLASSES_ROOT\cmcollectorsvideoplay\shell] >> "%TEMP_REG%"
echo. >> "%TEMP_REG%"
echo [-HKEY_CLASSES_ROOT\cmcollectorsvideoplay] >> "%TEMP_REG%"

echo 导入卸载注册表项...
reg import "%TEMP_REG%"
if %errorLevel% equ 0 (
    echo 成功导入注册表卸载文件!
) else (
    echo 警告: 导入注册表卸载文件失败!
)

echo 删除临时文件...
del "%TEMP_REG%"

echo.
echo 协议卸载尝试完成！
echo 如果仍然可以运行协议，请重启电脑后再试。
pause