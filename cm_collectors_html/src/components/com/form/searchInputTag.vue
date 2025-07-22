<template>
  <div class="el-input-tag el-input-tag__wrapper" :class="{ 'is-focused': isFocus }" tabindex="-1">
    <div class="el-input-tag__inner " :class="{ 'is-left-space': tags.length == 0 }">
      <!-- 标签区域 -->
      <slot name="tag"></slot>
      <el-tag v-for="(tag, index) in tags" :key="index" closable type="info" @close="removeTag(index)">
        {{ tag }}
      </el-tag>

      <!-- 输入框 -->
      <div class="el-input-tag__input-wrapper ">
        <input ref="inputRef" v-model="inputValue" @keyup.enter="addTag" @focus="isFocus = true" @blur="isFocus = false"
          class="el-input-tag__input" :placeholder="props.placeholder" style="min-width: 100px;" autocomplete="off"
          tabindex="0" />
      </div>
    </div>
    <div class="el-input-tag__suffix">
      <el-icon>
        <Search />
      </el-icon>
      <el-icon v-if="props.clearable" @click="clearTags">
        <CircleClose />
      </el-icon>
    </div>

  </div>
</template>
<script setup lang="ts">
import { ref, defineProps, defineEmits, watch } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { te } from 'element-plus/es/locales.mjs';

const props = defineProps<{
  modelValue: string[]
  clearable?: boolean
  placeholder?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string[]): void
  (e: 'add-tag', value: string): void
  (e: 'clear-click'): void
}>()

const inputRef = ref<HTMLInputElement | null>(null)
const inputValue = ref('')
const tags = ref<string[]>([...props.modelValue])
const isFocus = ref(false)

// 监听 modelValue 变化
watch(() => props.modelValue, (newVal) => {
  tags.value = [...newVal]
})

// 添加 Tag
const addTag = () => {
  const value = inputValue.value.trim()
  if (value && !tags.value.includes(value)) {
    tags.value.push(value)
    inputValue.value = ''
    emit('update:modelValue', [...tags.value])
    emit('add-tag', value)
  }
}

// 删除 Tag
const removeTag = (index: number) => {
  tags.value.splice(index, 1)
  emit('update:modelValue', [...tags.value])
}

// 清空所有 Tag
const clearTags = () => {
  tags.value = []
  emit('update:modelValue', [])
  emit('clear-click')
}

// 暴露方法（可选）
defineExpose({
  addTag,
  clearTags
})
</script>
<style lang="scss" scoped>
.el-input-tag__input::placeholder {
  font-size: 12px; // 设置你想要的字体大小
}
</style>
