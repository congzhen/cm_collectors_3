<script lang="ts">
  import { onMount } from "svelte";
  import {
    WindowMinimise,
    WindowToggleMaximise,
    Quit,
    WindowIsMaximised,
  } from "../wailsjs/runtime";
  import { GetURL } from "../wailsjs/go/main/App";

  let isMaximised = false;
  let title = "CM File Collectors";
  let showDragOverlay = false;
  let iframeSrc = "http://127.0.0.1:12345";

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

    // 监听鼠标按下事件，当在标题栏按下时显示覆盖层
    document.addEventListener("mousedown", handleMouseDown);
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

  function toggleMaximize() {
    WindowToggleMaximise();
    WindowIsMaximised().then((result) => (isMaximised = result));
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
        <button class="titlebar-button close-button" on:click={Quit}>×</button>
      </div>
    </div>

    <div class="content">
      <iframe id="webview" src={iframeSrc}></iframe>
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
  }

  #webview {
    width: 100%;
    height: 100%;
    border: none;
  }
</style>
