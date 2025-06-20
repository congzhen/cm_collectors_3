const LoadingService = (() => {
  let instance: HTMLDivElement | null = null;
  let stylesInjected = false;

  function createStyles() {
    if (stylesInjected) return;
    stylesInjected = true;
    const style = document.createElement('style');
    style.innerHTML = `
      /* 全屏加载层样式 */
      .loading-screen {
        position: fixed;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        background-color: rgba(0, 0, 0, 0.6); /* 黑色背景，透明度 0.9 */
        display: flex;
        justify-content: center;
        align-items: center;
        z-index: 9999;
      }

      /* 加载动画样式 */
      .loader {
        border: 4px solid #f3f3f3; /* 边框颜色 */
        border-top: 4px solid #3498db; /* 顶部边框颜色 */
        border-radius: 50%;
        width: 50px;
        height: 50px;
        animation: spin 2s linear infinite;
      }

      @keyframes spin {
        0% { transform: rotate(0deg); }
        100% { transform: rotate(360deg); }
      }
    `;
    document.head.appendChild(style);
  }

  function createLoading(): HTMLDivElement {
    const loadingScreen = document.createElement('div');
    loadingScreen.classList.add('loading-screen');

    const loader = document.createElement('div');
    loader.classList.add('loader');

    loadingScreen.appendChild(loader);
    return loadingScreen;
  }

  return {
    show(): HTMLDivElement {
      createStyles();
      if (!instance) {
        instance = createLoading();
        document.body.appendChild(instance);
      }
      return instance;
    },
    hide() {
      if (instance) {
        instance.remove();
        instance = null;
      }
    }
  };
})();

export { LoadingService };
