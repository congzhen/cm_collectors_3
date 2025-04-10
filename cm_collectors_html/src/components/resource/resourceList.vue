<template>
  <div class="resource-list">
    <div class="resource-list-index" v-if="props.showMode === E_resourceShowMode.Index">
      <ul>
        <li>1</li>
        <li>1</li>
        <li>1</li>
        <li>1</li>
        <li>1</li>
        <li>1</li>
        <li>1</li>
        <li>1</li>
        <li>1</li>
        <li>1</li>
      </ul>
    </div>
    <div class="resource-list-name" v-else>
      <ul>
        <li>
          <label>1.</label>
          <span>第十二个人 Den 12. mann HD MP4 2017 挪威 剧情 历史 战争 中文字幕.mp4</span>
        </li>
        <li>
          <label>2.</label>
          <span>宠物恐龙.mkv</span>
        </li>
        <li>
          <label>3.</label>
          <span>T-34坦克.mkv</span>
        </li>
        <li>
          <label>1.</label>
          <span>第十二个人 Den 12. mann HD MP4 2017 挪威 剧情 历史 战争 中文字幕.mp4</span>
        </li>
        <li>
          <label>2.</label>
          <span>宠物恐龙.mkv</span>
        </li>
        <li>
          <label>3.</label>
          <span>T-34坦克.mkv</span>
        </li>
        <li>
          <label>1.</label>
          <span>第十二个人 Den 12. mann HD MP4 2017 挪威 剧情 历史 战争 中文字幕.mp4</span>
        </li>
        <li>
          <label>2.</label>
          <span>宠物恐龙.mkv</span>
        </li>
        <li>
          <label>3.</label>
          <span>T-34坦克.mkv</span>
        </li>
      </ul>
    </div>
  </div>
</template>
<script lang="ts" setup>
import type { PropType } from 'vue'
import { E_resourceShowMode } from '@/dataType/app.dataType'
const props = defineProps({
  showMode: {
    type: String as PropType<(typeof E_resourceShowMode)[keyof typeof E_resourceShowMode]>,
    default: E_resourceShowMode.Index,
  },
})
</script>
<style lang="scss" scoped>
.resource-list {
  padding-bottom: 0.5em;
}
.resource-list-index {
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

.resource-list-name {
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
