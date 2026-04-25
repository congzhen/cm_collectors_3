<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import {
    WindowMinimise,
    WindowToggleMaximise,
    Quit,
    WindowIsMaximised,
  } from "../wailsjs/runtime";
  import { GetURL, OpenMultipleFilesDialog } from "../wailsjs/go/main/App";

  let isMaximised = false;
  let title = "CM File Collectors";
  let showDragOverlay = false;
  let iframeSrc = "http://127.0.0.1:12345";
  let zoomLevel = 1; // 缩放级别，默认为1 (100%)
  let webview: HTMLIFrameElement; // iframe 元素引用
  let showZoomControls = false; // 控制缩放按钮的显示状态

  onMount(async () => {
    // 组件挂载后检查窗口状态
    isMaximised = await WindowIsMaximised();

    // 获取从Go传递的URL参数
    try {
      const url = await GetURL();
      if (url) {
        iframeSrc = url;
      }
    } catch (e) {
      console.error("Failed to get URL from backend:", e);
    }

    // 从本地存储中读取缩放级别
    const savedZoomLevel = localStorage.getItem("cm-collectors-zoom");
    if (savedZoomLevel) {
      zoomLevel = parseFloat(savedZoomLevel);
      applyZoom();
    }

    // 监听鼠标按下事件，当在标题栏按下时显示覆盖层
    document.addEventListener("mousedown", handleMouseDown);

    // 确保在应用启动时调用代理设置
    setupIframeProxy();
  });

  function handleMouseDown(e: MouseEvent) {
    const target = e.target as HTMLElement;
    if (target.closest(".titlebar") && !target.closest(".titlebar-button")) {
      showDragOverlay = true;
    }
  }

  function handleMouseUp() {
    showDragOverlay = false;
  }

  // 切换缩放控制按钮的显示状态
  function toggleZoomControls() {
    showZoomControls = !showZoomControls;
  }

  // 增加缩放
  function zoomIn() {
    zoomLevel += 0.1;
    saveAndApplyZoom();
  }

  // 减少缩放
  function zoomOut() {
    if (zoomLevel > 0.1) {
      // 限制最小缩放为 10%
      zoomLevel -= 0.1;
      saveAndApplyZoom();
    }
  }

  // 重置缩放
  function resetZoom() {
    zoomLevel = 1;
    saveAndApplyZoom();
  }

  // 保存并应用缩放
  function saveAndApplyZoom() {
    // 保存缩放级别到本地存储
    localStorage.setItem("cm-collectors-zoom", zoomLevel.toString());
    applyZoom();
  }

  // 应用缩放
  function applyZoom() {
    if (webview) {
      webview.style.zoom = zoomLevel.toString();
    }
  }

  function toggleMaximize() {
    WindowToggleMaximise();
    WindowIsMaximised().then((result) => (isMaximised = result));
  }

  // 清理事件监听器
  onDestroy(() => {
    document.removeEventListener("mousedown", handleMouseDown);
  });

  /**
   * 监听来自 iframe 的消息并代理执行 Wails 原生功能
   */
  function setupIframeProxy() {
    console.log("[Host] Setting up iframe proxy listener...");
    window.addEventListener("message", async (event: MessageEvent) => {
      // 安全建议：在生产环境中验证 event.origin
      // if (event.origin !== "your-iframe-origin") return;

      const data = event.data;

      // 检查消息来源标识
      if (data && typeof data === "object" && data.source === "iframe-client") {
        const { action, payload, requestId } = data;
        console.log(`[Host] Received message from iframe: action=${action}, requestId=${requestId}`);

        try {
          let result: unknown;

          if (action === "handshake") {
            console.log("[Host] Handshake received, responding...");
            // 处理握手请求，简单返回 true 表示代理可用
            event.source?.postMessage(
              {
                source: "host-main-window",
                requestId,
                success: true,
                result: true,
              },
              { targetOrigin: "*" },
            );
            return; // 握手请求直接返回，不进入后续通用响应逻辑
          } else if (action === "wails.dialog.openMultipleFiles") {
            console.log("[Host] Opening multiple files dialog...");
            
            // 确保参数为字符串，避免 undefined 导致 Go 端类型断言失败或 panic
            const title = typeof payload.title === 'string' ? payload.title : "Select Files";
            const name = typeof payload.name === 'string' ? payload.name : "Files";
            const pattern = typeof payload.pattern === 'string' ? payload.pattern : "*.*";

            try {
              // 调用 Wails Go 方法
              // 请确保 app.go 中 OpenMultipleFilesDialog 的签名与此处一致
              // 例如: func (a *App) OpenMultipleFilesDialog(title, name, pattern string) ([]string, error)
              result = await OpenMultipleFilesDialog(title, name, pattern);
              console.log("[Host] Files selected:", result);
            } catch (goError) {
              console.error("[Host] Go method call failed:", goError);
              throw new Error(`Failed to open file dialog: ${goError}`);
            }
          } else {
            throw new Error(`Unknown action: ${action}`);
          }

          // 发送成功响应
          event.source?.postMessage(
            {
              source: "host-main-window",
              requestId,
              success: true,
              result,
            },
            { targetOrigin: "*" },
          ); // 生产环境请替换为具体的 iframe origin
        } catch (error: any) {
          console.error("[Host] Error handling iframe request:", error);
          // 发送错误响应
          event.source?.postMessage(
            {
              source: "host-main-window",
              requestId,
              success: false,
              error: error.message || "Internal Host Error",
            },
            { targetOrigin: "*" },
          );
        }
      }
    });
  }

</script>

<main>
  <div class="container">
    <div class="titlebar">
      <div class="titlebar-title">{title}</div>
      <div class="titlebar-controls">
        <button class="titlebar-button" on:click={WindowMinimise}>─</button>
        <button class="titlebar-button" on:click={toggleMaximize}>
          {isMaximised ? "❐" : "□"}
        </button>
        <!-- 添加缩放控制切换按钮，使用符号保持一致性 -->
        <button class="titlebar-button" on:click={toggleZoomControls}>±</button>
        <button class="titlebar-button close-button" on:click={Quit}>×</button>
      </div>
    </div>

    <div class="content">
      <iframe id="webview" bind:this={webview} src={iframeSrc}></iframe>

      <!-- 缩放控制按钮，通过按钮切换显示状态 -->
      {#if showZoomControls}
        <div class="zoom-controls">
          <button on:click={zoomIn}>+</button>
          <button on:click={zoomOut}>-</button>
          <button on:click={resetZoom}>↺</button>
          <span>{(zoomLevel * 100).toFixed(0)}%</span>
        </div>
      {/if}

      {#if showDragOverlay}
        <div
          class="drag-overlay"
          on:mouseup={handleMouseUp}
          on:mouseleave={handleMouseUp}
        ></div>
      {/if}
    </div>
  </div>
</main>

<style>
  .container {
    display: flex;
    flex-direction: column;
    height: 100vh;
    width: 100vw;
  }

  .titlebar {
    height: 20px;
    background: #1b1b1b;
    user-select: none;
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0 10px;
    color: white;
    font-family: "Segoe UI", sans-serif;
    font-size: 14px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
    --wails-draggable: drag;
    -webkit-user-select: none;
    position: relative;
    z-index: 100;
  }

  .titlebar-title {
    flex-grow: 1;
    text-align: center;
    margin: 0 40px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    --wails-draggable: drag;
  }

  .titlebar-controls {
    display: flex;
    height: 100%;
    --wails-draggable: no-drag;
  }

  .titlebar-button {
    background: none;
    border: none;
    color: white;
    width: 45px;
    height: 100%;
    cursor: pointer;
    font-size: 16px;
    display: flex;
    align-items: center;
    justify-content: center;
    --wails-draggable: no-drag;
    -webkit-user-select: none;
    user-select: none;
  }

  .titlebar-button:hover {
    background-color: rgba(255, 255, 255, 0.1);
  }

  .close-button:hover {
    background-color: #e81123 !important;
  }

  .content {
    flex: 1;
    background-color: #ffffff;
    position: relative;
    overflow: hidden;
  }

  .zoom-controls {
    position: absolute;
    top: 25px;
    right: 10px;
    z-index: 90;
    background: rgba(0, 0, 0, 0.7);
    padding: 5px;
    border-radius: 3px;
    color: white;
    display: flex;
    gap: 5px;
    align-items: center;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
  }

  .zoom-controls button {
    padding: 2px 8px;
    font-size: 12px;
    background: rgba(255, 255, 255, 0.2);
    border: none;
    color: white;
    border-radius: 2px;
    cursor: pointer;
  }

  .zoom-controls button:hover {
    background: rgba(255, 255, 255, 0.3);
  }

  #webview {
    width: 100%;
    height: 100%;
    border: none;
    transition: all 0.1s ease;
  }

  .drag-overlay {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    z-index: 1000;
  }
</style>

