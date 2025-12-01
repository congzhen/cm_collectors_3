// Electron 主进程文件
const { app, BrowserWindow, ipcMain } = require('electron');
const path = require('path');

// 获取命令行参数中的URL
function getUrlFromArgs() {
    const args = process.argv.slice(1);
    let url = 'http://127.0.0.1:12345';

    for (let i = 0; i < args.length; i++) {
        if (args[i] === '-url' && i + 1 < args.length) {
            return args[i + 1];
        } else if (args[i].startsWith('-url=')) {
            return args[i].substring(5);
        }
    }

    return url;
}

let mainWindow;

function createWindow() {
    const url = getUrlFromArgs();

    mainWindow = new BrowserWindow({
        width: 1280,
        height: 800,
        frame: false, // 无边框窗口
        icon: path.join(__dirname, 'resources/icon.ico'), // 添加图标路径
        webPreferences: {
            nodeIntegration: true,
            contextIsolation: false,
            enableRemoteModule: true,
        }
    });

    // 最大化窗口
    mainWindow.maximize();

    // 加载应用界面
    mainWindow.loadFile('frontend/index.html');

    // 当窗口加载完成，发送URL给渲染进程
    mainWindow.webContents.once('dom-ready', () => {
        mainWindow.webContents.send('set-url', url);
    });

    // 监听窗口控制事件
    ipcMain.on('window-minimize', () => {
        mainWindow.minimize();
    });

    ipcMain.on('window-maximize', () => {
        if (mainWindow.isMaximized()) {
            mainWindow.unmaximize();
        } else {
            mainWindow.maximize();
        }
    });

    ipcMain.on('window-close', () => {
        app.quit();
    });

    // 检查窗口是否最大化
    ipcMain.handle('window-is-maximized', () => {
        return mainWindow.isMaximized();
    });
}

app.whenReady().then(createWindow);

app.on('window-all-closed', () => {
    if (process.platform !== 'darwin') {
        app.quit();
    }
});

app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) {
        createWindow();
    }
});