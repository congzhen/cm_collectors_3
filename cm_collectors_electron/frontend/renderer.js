const { ipcRenderer } = require('electron');

let isMaximized = false;
let iframeSrc = 'http://127.0.0.1:12345';

// DOM元素
const minimizeBtn = document.getElementById('minimize-btn');
const maximizeBtn = document.getElementById('maximize-btn');
const closeBtn = document.getElementById('close-btn');
const webview = document.getElementById('webview');
const dragOverlay = document.getElementById('drag-overlay');
const titlebar = document.querySelector('.titlebar');

// 设置窗口控制事件监听器
minimizeBtn.addEventListener('click', () => {
    ipcRenderer.send('window-minimize');
});

maximizeBtn.addEventListener('click', () => {
    ipcRenderer.send('window-maximize');
});

closeBtn.addEventListener('click', () => {
    ipcRenderer.send('window-close');
});

// 标题栏拖拽处理
titlebar.addEventListener('mousedown', (e) => {
    if (!e.target.closest('.titlebar-button')) {
        dragOverlay.style.display = 'block';
    }
});

titlebar.addEventListener('mouseup', () => {
    dragOverlay.style.display = 'none';
});

titlebar.addEventListener('mouseleave', () => {
    dragOverlay.style.display = 'none';
});

// 接收主进程发送的URL
ipcRenderer.on('set-url', (event, url) => {
    if (url) {
        iframeSrc = url;
        webview.src = iframeSrc;
    }
});

// 添加键盘事件监听器实现缩放功能
document.addEventListener('keydown', (event) => {
    // 检查是否按下了 Ctrl 键和 +/- 键
    if (event.ctrlKey) {
        if (event.key === '+' || event.key === '=') {
            // 发送放大消息到主进程
            ipcRenderer.send('zoom-in');
            event.preventDefault();
        } else if (event.key === '-' || event.key === '_') {
            // 发送缩小消息到主进程
            ipcRenderer.send('zoom-out');
            event.preventDefault();
        } else if (event.key === '0') {
            // 发送重置缩放消息到主进程
            ipcRenderer.send('zoom-reset');
            event.preventDefault();
        }
    }
});

// 更新最大化按钮图标
async function updateMaximizeButton() {
    isMaximized = await ipcRenderer.invoke('window-is-maximized');
    maximizeBtn.textContent = isMaximized ? '❐' : '□';
}

// 定期更新最大化状态
setInterval(updateMaximizeButton, 1000);
updateMaximizeButton();

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    updateMaximizeButton();
});