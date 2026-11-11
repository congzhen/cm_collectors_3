<template>
  <div class="category-browse-fields" :class="{ 'is-compact': compact }">
    <div class="category-field">
      <span>{{ compact ? '开启分类浏览' : '分类浏览' }}</span
      ><el-switch v-model="config.categoryBrowseEnabled" aria-label="开启分类浏览" />
    </div>
    <template v-if="config.categoryBrowseEnabled || alwaysShow">
      <div class="category-field">
        <span>{{ compact ? '分类依据' : '浏览内容' }}</span
        ><el-select
          v-model="config.categoryBrowseSource"
          aria-label="分类浏览内容"
          :disabled="!config.categoryBrowseEnabled"
          :teleported="false"
        >
          <el-option label="演员" value="performer" />
          <el-option
            v-for="item in classes"
            :key="item.id"
            :label="item.name"
            :value="`tag:${item.id}`"
          />
        </el-select>
      </div>
      <div class="category-field">
        <span>分类排序</span
        ><el-select
          v-model="config.categoryBrowseSort"
          aria-label="分类排序"
          :disabled="!config.categoryBrowseEnabled"
          :teleported="false"
        >
          <el-option label="资源最多" value="count" />
          <el-option label="名称顺序" value="name" />
          <el-option label="默认顺序" value="default" />
        </el-select>
      </div>
      <div class="category-field">
        <span>每组资源上限</span
        ><el-input-number
          :key="config.categoryBrowseEnabled ? 'enabled' : 'disabled'"
          v-model="config.categoryBrowseGroupLimit"
          :disabled="!config.categoryBrowseEnabled"
          :min="0"
          :max="100"
          :precision="0"
          controls-position="right"
        />
      </div>
      <div class="category-field">
        <span>{{ compact ? '封面样式' : '合集封面样式' }}</span
        ><selectResourceGroupCoverStyle
          v-model="config.categoryBrowseCoverStyle" v-model:hide-single="config.categoryBrowseHideSingle" show-single-toggle
          :teleported="false"
          :disabled="!config.categoryBrowseEnabled"
        />
      </div>
      <div class="category-field">
        <span>{{ compact ? '展开样式' : '合集展开样式' }}</span
        ><selectResourceGroupExpandStyle
          v-model="config.categoryBrowseExpandStyle" v-model:adapt-cover-ratio="config.categoryBrowseAdaptCoverRatio"
          :teleported="false"
          :disabled="!config.categoryBrowseEnabled"
        />
      </div>
      <p v-if="!compact">
        上限 0 表示不限制、不拆组；1～100
        按上限拆组。跟随当前筛选，未分类置后。默认顺序为后添加的演员或标签优先。仅本库保存。
      </p>
      <p v-if="validationError && config.categoryBrowseEnabled" class="category-error">
        {{ validationError }}
      </p>
    </template>
    <p v-if="error && config.categoryBrowseEnabled" class="category-error">
      {{ error }} <el-button link @click="load">重试</el-button>
    </p>
  </div>
</template>
<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue'
import type { CategoryBrowseSettings } from '@/server/categoryBrowse.server'
import type { I_tagClass } from '@/dataType/tag.dataType'
import { tagServer } from '@/server/tag.server'
import selectResourceGroupCoverStyle from '@/components/com/form/selectResourceGroupCoverStyle.vue'
import selectResourceGroupExpandStyle from '@/components/com/form/selectResourceGroupExpandStyle.vue'
const props = defineProps<{
  config: CategoryBrowseSettings
  filesBasesId: string
  compact?: boolean
  alwaysShow?: boolean
}>()
const emit = defineEmits<{ ready: [] }>()
const classes = ref<I_tagClass[]>([]),
  error = ref('')
let version = 0
const validationError = ref('')
const loading = ref(false)
function validate() {
  validationError.value = ''
  if (!props.config.categoryBrowseEnabled) return true
  const limit = props.config.categoryBrowseGroupLimit
  if (!Number.isInteger(limit) || limit < 0 || limit > 100) {
    validationError.value = '每组资源上限请输入 0～100 的整数（0 表示不限制）'
  } else if (!['count', 'name', 'default'].includes(props.config.categoryBrowseSort)) {
    validationError.value = '请选择有效的分类排序'
  } else if (props.config.categoryBrowseSource !== 'performer' && loading.value) {
    validationError.value = '标签分类正在加载，请稍候'
  } else if (
    props.config.categoryBrowseSource !== 'performer' &&
    !classes.value.some((item) => `tag:${item.id}` === props.config.categoryBrowseSource)
  ) {
    validationError.value = '请选择当前文件库可用的标签分类，或重试加载'
  }
  return !validationError.value
}
defineExpose({ validate })
async function load() {
  const current = ++version
  classes.value = []
  error.value = ''
  loading.value = false
  if (!props.filesBasesId) return
  loading.value = true
  try {
    const result = await tagServer.tagClassListByFilesBasesId(props.filesBasesId)
    if (current !== version) return
    if (!result.status) throw new Error(result.msg)
    classes.value = result.data.filter((item) => item.status)
  } catch {
    if (current === version) error.value = '标签分类加载失败'
  } finally {
    if (current === version) {
      loading.value = false
      if (validationError.value) validate()
      emit('ready')
    }
  }
}
watch(() => props.filesBasesId, load, { immediate: true })
onBeforeUnmount(() => {
  version++
})
</script>
<style scoped>
.category-browse-fields {
  width: 100%;
  min-width: 0;
}
.category-field {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin: 10px 0;
}
.category-field:first-child {
  margin-top: 0;
}
.category-field > span {
  white-space: nowrap;
  font-size: 13px;
  color: var(--el-text-color-regular);
}
.category-field :deep(.el-select),
.category-field :deep(.el-input-number) {
  width: 160px;
  min-width: 0;
}
p {
  color: var(--el-text-color-secondary);
  font-size: 12px;
  line-height: 1.6;
  margin: 8px 0;
}
.category-error {
  color: var(--el-color-danger);
}
.is-compact .category-field {
  display: grid;
  grid-template-columns: 120px minmax(0, 1fr);
  gap: 0;
  margin: 0 0 16px;
  min-height: 32px;
}
.is-compact .category-field :deep(.el-select) {
  width: 100%;
}
.is-compact .category-field :deep(.el-input-number) {
  width: 110px;
}
.is-compact .category-field :deep(.el-switch) {
  justify-self: start;
}
</style>
