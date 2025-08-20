import type { App, Directive } from 'vue'
import { appStoreData } from '@/storeData/app.storeData'
import { watch } from 'vue'

/**
 * v-admin 指令
 * 根据用户是否为管理员来控制元素的显示
 *
 * 用法:
 * <div v-admin>只有管理员可见</div>
 * <div v-admin="false">只有非管理员可见</div>
 */

interface AdminHTMLElement extends HTMLElement {
  _unwatch?: () => void
}

const admin: Directive<HTMLElement, boolean | undefined> = {
  mounted(el) {
    const store = appStoreData()
    if (!store.isAdminLogin || store.isAdminLoginStatus) {
      // 显示元素
      el.style.display = ''
    } else {
      // 隐藏元素
      el.style.display = 'none'
    }
    // 监听store变化
    const adminEl = el as AdminHTMLElement
    adminEl._unwatch = watch(() => [store.isAdminLogin, store.isAdminLoginStatus], () => {
      if (!store.isAdminLogin || store.isAdminLoginStatus) {
        // 显示元素
        el.style.display = ''
      } else {
        // 隐藏元素
        el.style.display = 'none'
      }
    })
  },

  updated(el) {
    const store = appStoreData()
    if (!store.isAdminLogin || store.isAdminLoginStatus) {
      // 显示元素
      el.style.display = ''
    } else {
      // 隐藏元素
      el.style.display = 'none'
    }
  },

  unmounted(el) {
    // 清理监听
    const adminEl = el as AdminHTMLElement
    if (adminEl._unwatch) {
      adminEl._unwatch()
    }
  }
}

// 批量注册指令
export default {
  install(app: App) {
    app.directive('admin', admin)
  }
}
