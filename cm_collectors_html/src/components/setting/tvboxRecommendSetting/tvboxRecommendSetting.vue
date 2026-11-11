<template>
  <div class="tvbox-recommend-setting">
    <div class="toolbar">
      <span class="tip">拖拽调整排序，排在前面的资源将优先展示在 TVBox 首页推荐栏。</span>
      <el-button size="small" @click="loadList">刷新</el-button>
    </div>
    <div class="cache-tip">
      <el-icon><InfoFilled /></el-icon>
      修改推荐列表后，TVBox 可能需要手动清理缓存才能看到更新（长按站点 → 清除缓存）。
    </div>
    <div v-if="list.length === 0" class="empty-tip">暂无推荐内容，请在资源右键菜单中添加。</div>
    <draggable
      v-else
      class="recommend-list"
      :list="list"
      item-key="id"
      handle=".drag-handle"
      @end="onDragEnd"
    >
      <template #item="{ element }">
        <div class="recommend-item">
          <el-icon class="drag-handle"><Rank /></el-icon>
          <img
            class="cover"
            :src="getResourceCoverPoster(element.resource)"
            :alt="element.resource.title"
          />
          <span class="title">{{ element.resource.title }}</span>
          <el-button
            type="danger"
            text
            size="small"
            :loading="deleting === element.id"
            @click="deleteItem(element)"
          >
            <el-icon><Delete /></el-icon>
          </el-button>
        </div>
      </template>
    </draggable>
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import draggable from 'vuedraggable'
import { ElMessage } from 'element-plus'
import { Rank, Delete, InfoFilled } from '@element-plus/icons-vue'
import { tvboxRecommendServer } from '@/server/tvboxRecommend.server'
import type { I_tvboxRecommend } from '@/dataType/tvboxRecommend.dataType'
import { getResourceCoverPoster } from '@/common/photo'
import { debounce } from '@/assets/debounce'

const list = ref<I_tvboxRecommend[]>([])
const deleting = ref('')

const loadList = async () => {
  const result = await tvboxRecommendServer.list()
  if (result.status) {
    list.value = result.data
  } else {
    ElMessage.error(result.msg || '加载失败')
  }
}

const deleteItem = async (item: I_tvboxRecommend) => {
  deleting.value = item.id
  const result = await tvboxRecommendServer.delete(item.id)
  deleting.value = ''
  if (result.status) {
    list.value = list.value.filter((r) => r.id !== item.id)
    ElMessage.success('已移除')
  } else {
    ElMessage.error(result.msg || '删除失败')
  }
}

const onDragEnd = debounce(async () => {
  const total = list.value.length
  const sortItems = list.value.map((item, index) => ({ id: item.id, sort: total - 1 - index }))
  const result = await tvboxRecommendServer.updateSort(sortItems)
  if (!result.status) {
    ElMessage.error(result.msg || '排序保存失败')
  }
}, 600)

onMounted(loadList)
</script>

<style lang="scss" scoped>
@use '../settingsModules' as modules;

.tvbox-recommend-setting {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 10px;
  overflow: hidden;

  .toolbar {
    display: flex;
    align-items: center;
    gap: 10px;

    .tip {
      font-size: 12px;
      color: var(--el-text-color-secondary);
      flex: 1;
    }
  }

  .cache-tip {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    color: var(--el-color-warning);
    background: var(--el-color-warning-light-9);
    border-radius: 4px;
    padding: 6px 10px;
  }

  .empty-tip {
    color: var(--el-text-color-secondary);
    font-size: 13px;
    padding: 20px 0;
    text-align: center;
  }

  .recommend-list {
    flex: 1;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 6px;

    .recommend-item {
      display: flex;
      align-items: center;
      gap: 10px;
      padding: 6px 10px;
      border-radius: 6px;
      background: var(--el-fill-color-light);

      .drag-handle {
        cursor: move;
        color: var(--el-text-color-secondary);
        flex-shrink: 0;
      }

      .cover {
        width: 40px;
        height: 56px;
        object-fit: cover;
        border-radius: 4px;
        background: var(--el-fill-color);
        flex-shrink: 0;
      }

      .title {
        flex: 1;
        font-size: 13px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
    }
  }
}

.tvbox-recommend-setting {
  gap: 12px;
  .toolbar { @include modules.panel; padding: 8px 16px; flex-shrink: 0; }
  .cache-tip { color: var(--el-text-color-secondary); background: var(--el-fill-color-light); line-height: 1.7; flex-shrink: 0; }
  .recommend-list { min-height: 0; gap: 10px; .recommend-item { padding: 8px 16px; background: var(--el-bg-color); border: 1px solid var(--el-border-color-lighter); border-radius: 8px; } }
  .empty-tip { @include modules.panel; padding: 48px 20px; }
}

.tvbox-recommend-setting { @include modules.frame; }

.tvbox-recommend-setting {
  gap: 8px;
  .toolbar { padding: 8px 16px; }
  .cache-tip { padding: 6px 12px; font-size: 12px; }
  .recommend-list { gap: 0; border: 1px solid var(--el-border-color-lighter); border-radius: 8px; background: var(--el-bg-color); counter-reset: recommendations; }
  .recommend-list .recommend-item { flex-shrink: 0; gap: 12px; padding: 8px 12px; border: 0; border-bottom: 1px solid var(--el-border-color-extra-light); border-radius: 0; background: transparent; }
  .recommend-list .recommend-item:last-child { border-bottom: 0; }
  .recommend-list .recommend-item::before { counter-increment: recommendations; content: counter(recommendations, decimal-leading-zero); min-width: 24px; font-size: 12px; color: var(--el-text-color-secondary); font-variant-numeric: tabular-nums; }
  .recommend-list .recommend-item .cover { width: 32px; height: 44px; }
  .recommend-list .drag-handle { padding: 8px 4px; cursor: grab; }
  .recommend-list .title { font-size: 14px; }
  .recommend-list .recommend-item:hover { background: var(--el-fill-color-light); }
}
</style>
