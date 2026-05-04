import { computed } from 'vue'
import { appStoreData } from '@/storeData/app.storeData'
import { appDataServer } from '@/server/app.server'
import type { T_homeMode } from '@/dataType/app.dataType'

export type { T_homeMode }

export interface I_homeModeOption {
  label: string
  value: T_homeMode
}

export const homeModeOptions: I_homeModeOption[] = [
  { label: '经典布局', value: 'classic' },
  { label: '画室布局', value: 'studio' },
]

export const useHomeMode = () => {
  const store = appStoreData()

  const currentHomeMode = computed<T_homeMode>(() => {
    const mode = store.appConfig?.homeMode
    if (mode && homeModeOptions.some(item => item.value == mode)) {
      return mode
    }
    return 'classic'
  })

  const currentHomeModeLabel = computed(() => {
    return homeModeOptions.find(item => item.value == currentHomeMode.value)?.label || ''
  })

  const setHomeMode = async (mode: T_homeMode) => {
    if (store.appConfig) {
      store.appConfig.homeMode = mode
    }
    try {
      const configResult = await appDataServer.getAppConfig()
      if (configResult.status) {
        configResult.data.homeMode = mode
        await appDataServer.setAppConfig(configResult.data)
      }
    } catch (error) {
      console.log(error)
    }
  }

  return {
    currentHomeMode,
    currentHomeModeLabel,
    homeModeOptions,
    setHomeMode,
  }
}
