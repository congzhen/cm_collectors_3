// 全局菜单管理器
class ContextMenuManager {
  private static instance: ContextMenuManager
  private openMenu: (() => void) | null = null

  private constructor() { }

  static getInstance(): ContextMenuManager {
    if (!ContextMenuManager.instance) {
      ContextMenuManager.instance = new ContextMenuManager()
    }
    return ContextMenuManager.instance
  }

  registerMenu(closeFunction: () => void) {
    // 如果有已打开的菜单，先关闭它
    if (this.openMenu) {
      this.openMenu()
    }

    // 注册新的菜单关闭函数
    this.openMenu = closeFunction
  }

  unregisterMenu() {
    this.openMenu = null
  }
}

export default ContextMenuManager.getInstance()
