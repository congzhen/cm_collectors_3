@echo off
REM 设置代码页为 UTF-8 (65001) 以支持中文显示
chcp 65001 >nul
setlocal enabledelayedexpansion

REM 获取当前目录（根目录）
set ROOT_DIR=%cd%
set BUILD_DIR=!ROOT_DIR!\build

echo ========================================
echo CM Collectors 3 构建工具
echo ========================================
echo.

REM 初始化选择列表
set "SELECTED_ITEMS="
set "BUILD_FRONTEND=0"
set "BUILD_SERVER_WIN_TRAY_GUI=0"
set "BUILD_SERVER_WIN_TRAY=0"
set "BUILD_SERVER_WIN=0"
set "BUILD_SERVER_LINUX=0"
set "BUILD_WAILS=0"
set "BUILD_LAUNCHER=0"
set "BUILD_VIDEO_CALLER_WIN=0"
set "BUILD_VIDEO_CALLER_LINUX=0"
set "BUILD_SCRAPER_DEBUGGER=0"

:MENU
REM 清空输入变量，防止残留
set "INPUT="

echo 请选择要构建的项目 (输入数字，多个项目用空格分隔):
echo [1] cm_collectors_html (前端 - 优先构建)
echo [2] cm_collectors_server (Windows Tray + No Console GUI)
echo [3] cm_collectors_server (Windows Tray + Console)
echo [4] cm_collectors_server (Windows No Tray)
echo [5] cm_collectors_server (Linux)
echo [6] cm_collectors_wails
echo [7] cm_collectors_windows_launcher
echo [8] cm_collectors_video_caller (Windows)
echo [9] cm_collectors_video_caller (Linux)
echo [10] cm_collectors_scraper_debugger
echo.
echo 提示: 直接按回车键开始构建 (首次使用将构建默认项目: 前端+Linux服务端+Windows托盘版)
echo.

set /p INPUT="请输入选择: "

REM 如果未输入，检查是否有之前的选择
if "%INPUT%"=="" (
    REM 如果没有任何选择，使用默认值
    if "!BUILD_FRONTEND!!BUILD_SERVER_WIN_TRAY_GUI!!BUILD_SERVER_LINUX!"=="000" (
        set "BUILD_FRONTEND=1"
        set "BUILD_SERVER_LINUX=1"
        set "BUILD_SERVER_WIN_TRAY_GUI=1"
        echo 未检测到选择，使用默认构建配置: 前端 + Linux服务端 + Windows托盘版
    )
    REM 无论是否有默认值或已有选择，直接回车都应开始构建
    goto START_BUILD
) else (
    REM 解析输入
    set "VALID_INPUT=0"
    for %%i in (%INPUT%) do (
        set "ITEM=%%i"
        if "!ITEM!"=="1" (set "BUILD_FRONTEND=1" & set "VALID_INPUT=1")
        if "!ITEM!"=="2" (set "BUILD_SERVER_WIN_TRAY_GUI=1" & set "VALID_INPUT=1")
        if "!ITEM!"=="3" (set "BUILD_SERVER_WIN_TRAY=1" & set "VALID_INPUT=1")
        if "!ITEM!"=="4" (set "BUILD_SERVER_WIN=1" & set "VALID_INPUT=1")
        if "!ITEM!"=="5" (set "BUILD_SERVER_LINUX=1" & set "VALID_INPUT=1")
        if "!ITEM!"=="6" (set "BUILD_WAILS=1" & set "VALID_INPUT=1")
        if "!ITEM!"=="7" (set "BUILD_LAUNCHER=1" & set "VALID_INPUT=1")
        if "!ITEM!"=="8" (set "BUILD_VIDEO_CALLER_WIN=1" & set "VALID_INPUT=1")
        if "!ITEM!"=="9" (set "BUILD_VIDEO_CALLER_LINUX=1" & set "VALID_INPUT=1")
        if "!ITEM!"=="10" (set "BUILD_SCRAPER_DEBUGGER=1" & set "VALID_INPUT=1")
    )

    REM 如果输入无效，提示并返回菜单
    if "!VALID_INPUT!"=="0" (
        echo.
        echo 警告: 未识别的选择 "%INPUT%"，请重新输入。
        echo.
        goto MENU
    )

    echo.
    echo 已选择项目，按回车键确认并开始构建，或继续输入其他项目编号...
    set /p CONFIRM="确认构建? (直接回车确认, 输入 'c' 继续选择): "
    if /i "!CONFIRM!"=="c" goto MENU


    goto START_BUILD
)




:START_BUILD
:: 输出变量状态，便于调试
echo.
echo [调试] 当前变量状态：
echo BUILD_FRONTEND=%BUILD_FRONTEND%
echo BUILD_SERVER_WIN_TRAY_GUI=%BUILD_SERVER_WIN_TRAY_GUI%
echo BUILD_SERVER_WIN_TRAY=%BUILD_SERVER_WIN_TRAY%
echo BUILD_SERVER_WIN=%BUILD_SERVER_WIN%
echo BUILD_SERVER_LINUX=%BUILD_SERVER_LINUX%
echo BUILD_WAILS=%BUILD_WAILS%
echo BUILD_LAUNCHER=%BUILD_LAUNCHER%
echo BUILD_VIDEO_CALLER_WIN=%BUILD_VIDEO_CALLER_WIN%
echo BUILD_VIDEO_CALLER_LINUX=%BUILD_VIDEO_CALLER_LINUX%
echo BUILD_SCRAPER_DEBUGGER=%BUILD_SCRAPER_DEBUGGER%
echo.
echo 进入 START_BUILD ...
echo.
echo ========================================
echo 开始构建选中的项目...
echo ========================================
echo 提示: 如果窗口即将关闭，请检查是否安装了 Node.js, Yarn 和 Go 环境。
echo.


REM 创建 build 目录
if not exist "!BUILD_DIR!" mkdir "!BUILD_DIR!"



REM --- 第一步：构建前端 (如果选中) ---

if "!BUILD_FRONTEND!"=="1" (    
    echo [前端] 构建中...
    if not exist "!ROOT_DIR!\cm_collectors_html" (
        echo [错误] 缺少 cm_collectors_html 目录
        goto END
    )
    pushd "!ROOT_DIR!\cm_collectors_html"
    call yarn build-server || (echo [错误] 前端构建失败 & popd & goto END)
    popd
    echo [前端] 构建完成
    pushd "!BUILD_DIR!"
)

REM --- 第二步：构建后端服务器 (根据选项) ---

REM 选项 2: Windows Tray + No Console GUI
if "!BUILD_SERVER_WIN_TRAY_GUI!"=="1" (
    echo [服务端Win Tray GUI] 构建中...
    pushd "!ROOT_DIR!\cm_collectors_server"
    set GOOS=windows
    set GOARCH=amd64
    go build -ldflags -H=windowsgui -tags tray -o "!BUILD_DIR!\cm_collectors_server.exe" . || (echo [错误] Win Tray GUI 构建失败 & popd & goto END)
    popd
    echo [服务端Win Tray GUI] 构建完成
    pushd "!BUILD_DIR!"
)

REM 选项 3: Windows Tray + Console
if "!BUILD_SERVER_WIN_TRAY!"=="1" (
    echo [服务端Win Tray Console] 构建中...
    pushd "!ROOT_DIR!\cm_collectors_server"
    set GOOS=windows
    set GOARCH=amd64
    go build -tags tray -o "!BUILD_DIR!\cm_collectors_server.exe" . || (echo [错误] Win Tray Console 构建失败 & popd & goto END)
    popd
    echo [服务端Win Tray Console] 构建完成
    pushd "!BUILD_DIR!"
)

REM 选项 4: Windows No Tray
if "!BUILD_SERVER_WIN!"=="1" (
    echo [服务端Win无托盘] 构建中...
    pushd "!ROOT_DIR!\cm_collectors_server"
    set GOOS=windows
    set GOARCH=amd64
    go build -o "!BUILD_DIR!\cm_collectors_server.exe" . || (echo [错误] Win无托盘 构建失败 & popd & goto END)
    popd
    echo [服务端Win无托盘] 构建完成
    pushd "!BUILD_DIR!"
)

REM 选项 5: Linux
if "!BUILD_SERVER_LINUX!"=="1" (
    echo [服务端Linux] 构建中...
    pushd "!ROOT_DIR!\cm_collectors_server"
    set GOOS=linux
    set GOARCH=amd64
    go build -o "!BUILD_DIR!\cm_collectors_server" . || (echo [错误] Linux 构建失败 & popd & goto END)
    popd
    echo [服务端Linux] 构建完成
    pushd "!BUILD_DIR!"
)

REM --- 其他子项目构建 ---

REM 选项 6: Wails
if "!BUILD_WAILS!"=="1" (
    echo [Wails] 构建中...
    pushd "!ROOT_DIR!\cm_collectors_wails"
    wails build -o ../../../build/cm_collectors_wails.exe || (echo [错误] Wails 构建失败 & popd & goto END)
    popd
    echo [Wails] 构建完成
    pushd "!BUILD_DIR!"
)

REM 选项 7: Windows Launcher
if "!BUILD_LAUNCHER!"=="1" (
    echo [启动器] 构建中...
    pushd "!ROOT_DIR!\cm_collectors_windows_launcher"
    go build -ldflags -H=windowsgui -o "!BUILD_DIR!\CM_launcher.exe" . || (echo [错误] 启动器构建失败 & popd & goto END)
    popd
    echo [启动器] 构建完成
    pushd "!BUILD_DIR!"
)

REM 选项 8: Video Caller Windows
if "!BUILD_VIDEO_CALLER_WIN!"=="1" (
    echo [云播Win] 构建中...
    pushd "!ROOT_DIR!\cm_collectors_video_caller"
    set GOOS=windows
    set GOARCH=amd64
    if not exist "!BUILD_DIR!\video_caller" mkdir "!BUILD_DIR!\video_caller"
    go build -ldflags -H=windowsgui -tags tray -o "!BUILD_DIR!\video_caller\cm_collectors_video_caller.exe" . || (echo [错误] 云播Win构建失败 & popd & goto END)
    popd
    echo [云播Win] 构建完成
    pushd "!BUILD_DIR!"
)

REM 选项 9: Video Caller Linux
if "!BUILD_VIDEO_CALLER_LINUX!"=="1" (
   echo [云播Linux] 构建中...
    pushd "!ROOT_DIR!\cm_collectors_video_caller"
    set GOOS=linux
    set GOARCH=amd64
    if not exist "!BUILD_DIR!\video_caller" mkdir "!BUILD_DIR!\video_caller"
    go build -o "!BUILD_DIR!\video_caller\cm_collectors_video_caller" . || (echo [错误] 云播Linux构建失败 & popd & goto END)
    popd
    echo [云播Linux] 构建完成
    pushd "!BUILD_DIR!"
)

REM 选项 10: Scraper Debugger
if "!BUILD_SCRAPER_DEBUGGER!"=="1" (
    echo [刮削调试器] 构建中...
    pushd "!ROOT_DIR!\cm_collectors_scraper_debugger"
    set GOOS=windows
    set GOARCH=amd64
    if not exist "!BUILD_DIR!\scraper_debugger" mkdir "!BUILD_DIR!\scraper_debugger"
    go build -o "!BUILD_DIR!\scraper_debugger\cm_collectors_scraper_debugger.exe" . || (echo [错误] 刮削调试器构建失败 & popd & goto END)
    popd
    echo [刮削调试器] 构建完成
    pushd "!BUILD_DIR!"
)



REM --- 询问是否复制资源文件 ---
set "COPY_SERVER_ASSETS=0"
set "COPY_VIDEO_CALLER_ASSETS=0"
set "COPY_SCRAPER_ASSETS=0"

if "!BUILD_SERVER_WIN_TRAY_GUI!"=="1" set "COPY_SERVER_ASSETS=1"
if "!BUILD_SERVER_WIN_TRAY!"=="1" set "COPY_SERVER_ASSETS=1"
if "!BUILD_SERVER_WIN!"=="1" set "COPY_SERVER_ASSETS=1"
if "!BUILD_SERVER_LINUX!"=="1" set "COPY_SERVER_ASSETS=1"

if "!BUILD_VIDEO_CALLER_WIN!"=="1" set "COPY_VIDEO_CALLER_ASSETS=1"
if "!BUILD_VIDEO_CALLER_LINUX!"=="1" set "COPY_VIDEO_CALLER_ASSETS=1"

if "!BUILD_SCRAPER_DEBUGGER!"=="1" set "COPY_SCRAPER_ASSETS=1"

REM 处理 Server 资源复制
if "!COPY_SERVER_ASSETS!"=="1" (
    echo.
    set /p COPY_CONFIRM="是否复制 cm_collectors_server 所需配置文件 (config.yaml, ffmpeg)? (Y/N): "
    if /i "!COPY_CONFIRM!"=="Y" (
        echo 正在复制 Server 资源文件...
        cd /d "!ROOT_DIR!\cm_collectors_server"
        call :CopyServerAssets
        cd /d "!ROOT_DIR!"
        echo Server 资源文件复制完成。
    ) else (
        echo 已跳过 Server 资源文件复制。
    )
)

REM 处理 Video Caller 资源复制
if "!COPY_VIDEO_CALLER_ASSETS!"=="1" (
    echo.
    set /p COPY_VC_CONFIRM="是否复制 cm_collectors_video_caller 所需配置文件 (config.json, scripts)? (Y/N): "
    if /i "!COPY_VC_CONFIRM!"=="Y" (
        echo 正在复制 Video Caller 资源文件...
        cd /d "!ROOT_DIR!\cm_collectors_video_caller"
        call :CopyVideoCallerAssets
        cd /d "!ROOT_DIR!"
        echo Video Caller 资源文件复制完成。
    ) else (
        echo 已跳过 Video Caller 资源文件复制。
    )
)

REM 处理 Scraper Debugger 资源复制
if "!COPY_SCRAPER_ASSETS!"=="1" (
    echo.
    set /p COPY_SD_CONFIRM="是否复制 cm_collectors_scraper_debugger 所需配置文件 (test.json, scraper dir)? (Y/N): "
    if /i "!COPY_SD_CONFIRM!"=="Y" (
        echo 正在复制 Scraper Debugger 资源文件...
        cd /d "!ROOT_DIR!\cm_collectors_scraper_debugger"
        call :CopyScraperAssets
        cd /d "!ROOT_DIR!"
        echo Scraper Debugger 资源文件复制完成。
    ) else (
        echo 已跳过 Scraper Debugger 资源文件复制。
    )
)

:: END 标签前 clear 屏并 pause，确保用户能看到所有输出
:END
echo.
echo ========== 构建流程结束，按任意键退出 ==========
echo.
echo ========================================
echo 所有构建任务已完成！
echo ========================================
echo.
echo 请按任意键退出...
pause >nul

REM 辅助函数：复制 Server 通用资源
:CopyServerAssets
copy /Y "config.yaml" "!BUILD_DIR!\" >nul 2>&1
if exist "ffmpeg" (
    robocopy "ffmpeg" "!BUILD_DIR!\ffmpeg" /E >nul 2>&1
)
exit /b

REM 辅助函数：复制 Video Caller 资源
:CopyVideoCallerAssets
if not exist "!BUILD_DIR!\video_caller" mkdir "!BUILD_DIR!\video_caller"
copy /Y "config.json" "!BUILD_DIR!\video_caller\" >nul 2>&1
copy /Y "setup_cm_video_caller.bat" "!BUILD_DIR!\video_caller\" >nul 2>&1
copy /Y "setup_cm_video_caller.sh" "!BUILD_DIR!\video_caller\" >nul 2>&1
copy /Y "uninstall_cm_video_caller.bat" "!BUILD_DIR!\video_caller\" >nul 2>&1
copy /Y "uninstall_cm_video_caller.sh" "!BUILD_DIR!\video_caller\" >nul 2>&1
exit /b

REM 辅助函数：复制 Scraper Debugger 资源
:CopyScraperAssets
if not exist "!BUILD_DIR!\scraper_debugger" mkdir "!BUILD_DIR!\scraper_debugger"
copy /Y "test.json" "!BUILD_DIR!\scraper_debugger\" >nul 2>&1
robocopy "scraper" "!BUILD_DIR!\scraper_debugger\scraper\" /E >nul 2>&1
exit /b
