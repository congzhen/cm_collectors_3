import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { appStoreData } from '@/storeData/app.storeData'
import { appDataServer } from '@/server/app.server'

export type T_theme = 'default' | 'bright'

export const applyTheme = (theme: string) => {
  const htmlElement = document.documentElement
  const isBright = theme === 'bright'
  htmlElement.classList.toggle('bright', isBright)
  htmlElement.classList.toggle('dark', !isBright)
}

export const useTheme = () => {
  const store = appStoreData()
  const isThemeSwitching = ref(false)

  const currentTheme = computed<T_theme>(() => store.appConfig?.theme === 'bright' ? 'bright' : 'default')
  const isBrightTheme = computed(() => currentTheme.value === 'bright')
  const themeToggleTitle = computed(() => isBrightTheme.value ? '切换为暗黑主题' : '切换为明亮主题')

  const setTheme = async (theme: T_theme) => {
    if (isThemeSwitching.value || theme === currentTheme.value) return true

    const previousTheme = currentTheme.value
    store.appConfig.theme = theme
    applyTheme(theme)
    isThemeSwitching.value = true

    try {
      const configResult = await appDataServer.getAppConfig()
      if (!configResult.status) throw new Error(configResult.msg)

      configResult.data.theme = theme
      const saveResult = await appDataServer.setAppConfig(configResult.data)
      if (!saveResult.status) throw new Error(saveResult.msg)
      return true
    } catch (error) {
      store.appConfig.theme = previousTheme
      applyTheme(previousTheme)
      ElMessage.error('主题设置保存失败，已恢复原主题')
      console.log(error)
      return false
    } finally {
      isThemeSwitching.value = false
    }
  }

  const toggleTheme = () => setTheme(isBrightTheme.value ? 'default' : 'bright')

  return {
    currentTheme,
    isBrightTheme,
    isThemeSwitching,
    themeToggleTitle,
    setTheme,
    toggleTheme,
  }
}
