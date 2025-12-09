import './style.css'
import './styleBright.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

import App from './App.vue'
import router from './router'

import { registerLang } from '@/language/app.lang.main'

// 导入自定义指令
import directives from './directives'

// 创建全局事件总线
import mitt from 'mitt'
export const eventBus = mitt()

const app = createApp(App)
registerLang(app)
app.use(createPinia())
app.use(router)
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}
app.use(ElementPlus).use(directives)
app.mount('#app')
