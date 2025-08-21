// 简单的语言包实现
import zhCn from './zhCn'

type Language = 'zhCn'
// eslint-disable-next-line @typescript-eslint/no-explicit-any
type LanguagePack = Record<string, string | Record<string, any>>

const languagePacks: Record<Language, LanguagePack> = {
  zhCn: zhCn as LanguagePack
}

let currentLanguage: Language = 'zhCn'

// 设置当前语言
export function setLanguage(lang: Language) {
  currentLanguage = lang
}

// 获取当前语言
export function getCurrentLanguage(): Language {
  return currentLanguage
}

// 获取翻译文本
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function t(key: string, params?: Record<string, any>): string {
  const pack = languagePacks[currentLanguage]

  // 处理嵌套键，例如 "sort.addTimeDesc"
  const keys = key.split('.')
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  let result: any = pack

  for (const k of keys) {
    if (result && typeof result === 'object' && result.hasOwnProperty(k)) {
      result = result[k]
    } else {
      // 如果找不到对应键，返回原始键
      return key
    }
  }

  // 如果结果是字符串且有参数，则替换参数
  if (typeof result === 'string' && params) {
    return result.replace(/\{(\w+)\}/g, (match, key) => {
      return params[key] !== undefined ? params[key] : match
    })
  }

  // 如果结果是对象，转换为JSON字符串（便于调试）
  if (typeof result === 'object') {
    return JSON.stringify(result)
  }

  return typeof result === 'string' ? result : key
}

// 注册函数，用于在Vue实例上注册$t函数
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function registerLang(app: any) {
  // 注册全局属性
  app.config.globalProperties.$t = t
  // 注册全局方法
  app.provide('$t', t)
}

// 导出语言包以供直接使用
export { zhCn }
export default {
  t,
  setLanguage,
  getCurrentLanguage
}
