// vue-cropper-wrapper.ts
// 这个文件用于包装vue-cropper组件，避免TypeScript类型检查错误

// @ts-expect-error 忽略类型检查错误
import { VueCropper } from 'vue-cropper/dist/vue-cropper.es.js'

import 'vue-cropper/dist/index.css'

export { VueCropper }
