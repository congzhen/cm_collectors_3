<template>
  <div class="resourceDramaSeries-list">
    <div class="resourceDramaSeries-list-index" v-if="props.showMode === E_detailsDramaSeriesMode.digit">
      <ul>
        <li v-for="(item, key) in props.dramaSeries" :key="key" @click="emits('playResourceDramaSeries', item)">
          {{ (key + 1) }}
        </li>
      </ul>
    </div>
    <div class="resourceDramaSeries-list-name" v-else>
      <ul>
        <li v-for="(item, key) in props.dramaSeries" :key="key" @click="emits('playResourceDramaSeries', item)">
          <label>{{ (key + 1) }}.</label>
          <span>{{ getFinalPathSegment(item.src) }}</span>
        </li>
      </ul>
    </div>
  </div>
</template>
<script lang="ts" setup>
import type { PropType } from 'vue'
import { E_detailsDramaSeriesMode } from '@/dataType/app.dataType'
import type { I_resourceDramaSeries } from '@/dataType/resource.dataType';
import { getFinalPathSegment } from '@/assets/tool'
const props = defineProps({
  showMode: {
    type: String as PropType<(typeof E_detailsDramaSeriesMode)[keyof typeof E_detailsDramaSeriesMode]>,
    default: E_detailsDramaSeriesMode.fileName,
  },
  dramaSeries: {
    type: Array as PropType<I_resourceDramaSeries[]>,
    required: true,
  }
})

const emits = defineEmits(['playResourceDramaSeries']);


</script>
<style lang="scss" scoped>
.resourceDramaSeries-list {
  padding-bottom: 0.5em;
}

.resourceDramaSeries-list-index {
  ul {
    width: 100%;
    list-style-type: none;
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 5px;

    li {
      width: 17%;
      height: 1.2em;
      line-height: 1.2em;
      font-weight: 500;
      text-align: center;
      padding: 0.5em 0;
      border: 1px solid #a8abb2;
      border-radius: 4px;
      cursor: pointer;
      user-select: none;
      /* 新增过渡动画 */
      transition:
        background-color 0.2s ease,
        box-shadow 0.2s ease;

      /* 鼠标悬停时的高亮效果 */
      &:hover {
        color: var(--el-color-primary);
        background-color: var(--el-color-primary-light-9);
        border-color: var(--el-color-primary);
      }
    }
  }
}

.resourceDramaSeries-list-name {
  ul {
    width: calc(100% - 0.4em);
    margin: 0 0.2em;
    list-style-type: none;

    li {
      width: calc(100% - 2em);
      line-height: 1.2em;
      font-weight: 500;
      font-style: italic;
      padding: 0.5em 1em;
      border-bottom: 1px dotted rgba(168, 171, 178, 0.5);
      cursor: pointer;
      user-select: none;
      display: flex;
      /* 新增过渡动画 */
      transition:
        background-color 0.2s ease,
        box-shadow 0.2s ease;

      /* 鼠标悬停时的高亮效果 */
      &:hover {
        color: var(--el-color-primary);
        background-color: var(--el-color-primary-light-9);
        border-color: var(--el-color-primary);
      }

      label {
        flex-shrink: 0;
        padding-right: 0.8em;
      }

      span {
        flex-grow: 1;
      }
    }
  }
}
</style>
