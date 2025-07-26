import { createI18n } from 'vue-i18n'
import zhCn from './zhCn'
const i18nMessage = {
  zhCn: zhCn
}
const i18n = createI18n({
  globalInjection: true,
  locale: 'zhCn',
  messages: i18nMessage,
  legacy: false
})
export default i18n
