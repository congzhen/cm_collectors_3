/**
 * 运行时桥接工具类
 * 用于在 iframe 或子页面中安全地调用宿主环境（如 Wails, Electron 等）的功能
 *
 * 注意：在 iframe 环境中，通常无法直接访问父窗口的 window.__WAILS__。
 * 因此，本桥接层主要依赖 postMessage 与父窗口通信，由父窗口代理执行原生调用。
 */

// 定义更具体的接口以替代 any

// Wails v2 全局对象结构示例 (根据实际项目调整)
export interface WailsGlobal {
  Go: Record<string, (...args: unknown[]) => Promise<unknown>>; // 假设 Go 对象包含异步方法
  Events: {
    On: (eventName: string, callback: (...args: unknown[]) => void) => void;
    Off: (eventName: string, callback: (...args: unknown[]) => void) => void;
    Emit: (eventName: string, ...args: unknown[]) => void;
  };
  // 可以添加其他已知属性
}

// Electron 全局对象结构示例 (根据实际 preload 脚本调整)
export interface ElectronGlobal {
  ipcRenderer?: {
    send: (channel: string, ...args: unknown[]) => void;
    on: (channel: string, listener: (event: unknown, ...args: unknown[]) => void) => void;
    removeListener: (channel: string, listener: (event: unknown, ...args: unknown[]) => void) => void;
    invoke: (channel: string, ...args: unknown[]) => Promise<unknown>;
  };
  shell?: {
    openExternal: (url: string) => Promise<void>;
  };
  // 可以添加其他通过 contextBridge 暴露的 API
}

// 定义全局变量的类型接口
declare global {
  interface Window {
    __WAILS__?: WailsGlobal; // Wails v2 通常注入此对象
    __ELECTRON__?: ElectronGlobal; // Electron 可能存在的标识
    go?: WailsGlobal['Go'];       // 某些配置下可能直接暴露 go 对象
    electron?: ElectronGlobal; // Electron preload 可能暴露的对象
  }
}


export async function runRuntimeBridge(): Promise<boolean> {
  // 先检测环境是否支持
  const isNative = await isNativeAppEnvironment();

  if (isNative) {
    // 如果支持，则初始化监听器
    listenToHostParent();
    console.log('[RuntimeBridge] Native environment detected, listener active.');
  } else {
    console.warn('[RuntimeBridge] Native environment not detected, listener skipped.');
  }

  return isNative;
}

/**
 * 判断当前环境是否支持原生应用功能调用
 *
 * 在 iframe 中，我们无法直接检测 window.__WAILS__，因为它位于父窗口。
 * 这里我们通过向父窗口发送一个握手消息来确认连通性。
 */
export function isNativeAppEnvironment(): Promise<boolean> {
  // 在 iframe 中，通过 postMessage 握手检测
  return checkHostAvailability();
}

/**
 * 内部辅助函数：通过握手消息检测宿主环境可用性
 */
function checkHostAvailability(): Promise<boolean> {
  return new Promise((resolve) => {
    if (!window.parent) {
      resolve(false);
      return;
    }

    const requestId = generateRequestId();
    const timeoutMs = 20000; // 较短的超时时间用于检测

    const timeoutId = window.setTimeout(() => {
      cleanup();
      resolve(false);
    }, timeoutMs);

    const messageHandler = (event: MessageEvent<unknown>) => {
      if (
        event.data &&
        typeof event.data === 'object' &&
        'source' in event.data &&
        (event.data as { source: unknown }).source === 'host-main-window' &&
        (event.data as { requestId?: string }).requestId === requestId
      ) {
        cleanup();
        resolve(true);
      }
    };

    const cleanup = () => {
      clearTimeout(timeoutId);
      window.removeEventListener('message', messageHandler);
    };

    window.addEventListener('message', messageHandler);

    // 发送握手消息
    window.parent.postMessage({
      source: 'iframe-client',
      requestId,
      action: 'handshake', // 定义一个特殊的握手动作
      payload: null,
      timestamp: Date.now()
    }, '*');
  });
}

/**
 * 生成唯一的请求 ID
 */
function generateRequestId(): string {
  return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
}

// 存储待处理的请求回调
const pendingRequests = new Map<string, {
  resolve: (value: unknown) => void;
  reject: (reason?: unknown) => void;
  timeoutId: number;
}>();

/**
 * 向父窗口发送异步请求并等待响应
 * 适用于 iframe 嵌入场景，代理调用 Wails/Electron 功能
 * @param action 动作名称，例如 'wails.call', 'electron.invoke'
 * @param payload 数据负载，包含具体调用的方法名和参数
 * @param timeoutMs 超时时间，默认 60000ms
 * @returns Promise<unknown>
 */
export function sendToHostParentAsync(action: string, payload: unknown, timeoutMs: number = 60000): Promise<unknown> {
  return new Promise((resolve, reject) => {
    if (!window.parent) {
      reject(new Error('No parent window found'));
      return;
    }

    const requestId = generateRequestId();

    // 设置超时
    const timeoutId = window.setTimeout(() => {
      if (pendingRequests.has(requestId)) {
        pendingRequests.delete(requestId);
        reject(new Error(`Request to host parent timed out: ${action}`));
      }
    }, timeoutMs);

    // 存储回调
    pendingRequests.set(requestId, { resolve, reject, timeoutId });

    // 发送消息
    window.parent.postMessage({
      source: 'iframe-client',
      requestId, // 添加 requestId 以便匹配响应
      action,
      payload,
      timestamp: Date.now()
    }, '*'); // 在生产环境中建议将 '*' 替换为具体的 targetOrigin
  });
}

/**
 * 监听来自宿主主窗口的响应
 * 自动匹配 requestId 并解析 Promise
 * @returns 一个用于取消监听的函数
 */
export function listenToHostParent(callback?: (data: unknown) => void): () => void {
  const messageHandler = (event: MessageEvent<unknown>) => {
    // 可选：验证 event.origin 以确保安全性
    // if (event.origin !== "expected-origin") return;

    // 类型守卫：确保 data 是对象且具有 source 属性
    if (
      event.data &&
      typeof event.data === 'object' &&
      'source' in event.data &&
      (event.data as { source: unknown }).source === 'host-main-window'
    ) {
      const responseData = event.data as { requestId?: string; success?: boolean; result?: unknown; error?: string };

      // 如果存在 requestId，说明是异步请求的响应
      if (responseData.requestId && pendingRequests.has(responseData.requestId)) {
        const request = pendingRequests.get(responseData.requestId)!;
        clearTimeout(request.timeoutId);
        pendingRequests.delete(responseData.requestId);

        if (responseData.success) {
          request.resolve(responseData.result);
        } else {
          request.reject(new Error(responseData.error || 'Unknown error from host'));
        }
      } else {
        // 否则，触发旧版的通用回调（如果提供）
        if (callback) {
          callback(event.data);
        }
      }
    }
  };

  window.addEventListener('message', messageHandler);

  // 返回取消监听的函数
  return () => {
    window.removeEventListener('message', messageHandler);
  };
}


/**
 * 在 iframe 中调用 Wails 的多选文件对话框
 * 注意：这需要父窗口监听 'wails.dialog.openMultipleFiles' 消息并代理调用 runtime.OpenMultipleFilesDialog
 * @param title 对话框标题
 * @param filters 文件过滤器，例如 [{ Name: "Images", Pattern: "*.jpg;*.png" }]
 * @returns Promise<string[]> 选中的文件路径数组
 */
export async function openMultipleFilesDialog(title?: string, name?: string, pattern?: string): Promise<string[]> {
  // 如果在顶层窗口且直接有 Wails 运行时，可以直接调用（如果暴露了 runtime）
  // 注意：Wails v2 的 runtime 通常不直接挂在 window 上，而是通过 Go 绑定或特定 API
  // 这里主要关注 iframe 场景，统一使用 postMessage 代理以保持一致性
  return sendToHostParentAsync('wails.dialog.openMultipleFiles', {
    title,
    name,
    pattern
  }) as Promise<string[]>;
}

export async function openDirectoryDialog(title?: string): Promise<string> {
  return sendToHostParentAsync('wails.dialog.openDirectory', {
    title
  }) as Promise<string>;
}
