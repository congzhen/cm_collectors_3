<template>
  <div class="mobile-header">
    <el-button icon="ArrowLeft" @click="goBack" circle size="small" />
    <div class="title" :title="title">
      {{ props.title }}
    </div>
    <el-button icon="More" circle size="small" @click="showMenu = !showMenu" v-if="props.showMenuButton" />

    <!-- 菜单弹窗 -->
    <div v-if="showMenu" class="menu-overlay" @click="showMenu = false">
      <div class="menu-popup" @click.stop>
        <div class="menu-item" @click="goBack">返回</div>
        <div class="menu-item" @click="goToHome">首页</div>
        <div class="menu-item" @click="showMenu = false">取消</div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref } from "vue";
import { useRouter } from 'vue-router';

const router = useRouter();
const showMenu = ref(false);

const props = defineProps({
  title: {
    type: String,
    default: ''
  },
  showMenuButton: {
    type: Boolean,
    default: true
  }
});

// 返回上一页
const goBack = () => {
  router.go(-1);
};

// 返回首页
const goToHome = () => {
  router.push('/mobile');
};
</script>

<style lang="scss" scoped>
.mobile-header {
  --mobile-header-bg: #1b2024;
  --mobile-header-surface: #262c31;
  --mobile-header-border: rgba(255, 255, 255, 0.12);
  --mobile-header-text: #edf1f3;

  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 52px;
  padding: calc(6px + env(safe-area-inset-top)) 12px 6px;
  box-sizing: border-box;
  color: var(--mobile-header-text);
  background-color: var(--mobile-header-bg);
  border-bottom: 1px solid var(--mobile-header-border);
  position: sticky;
  top: 0;
  z-index: 100;

  > .el-button {
    width: 36px;
    height: 36px;
    color: var(--mobile-header-text);
    background: var(--mobile-header-surface);
    border-color: var(--mobile-header-border);
  }

  .title {
    flex: 1;
    text-align: center;
    font-size: 16px;
    font-weight: 500;
    padding: 0 10px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    color: var(--mobile-header-text);
  }

  .menu-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.5);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 1000;

    .menu-popup {
      color: var(--mobile-header-text);
      background-color: var(--mobile-header-bg);
      border: 1px solid var(--mobile-header-border);
      border-radius: 8px;
      width: 200px;

      .menu-item {
        padding: 15px;
        text-align: center;
        border-bottom: 1px solid var(--mobile-header-border);
        cursor: pointer;
        color: var(--mobile-header-text);

        &:last-child {
          border-bottom: none;
        }

        &:active {
          background-color: var(--mobile-header-surface);
        }
      }
    }
  }
}

:global(html.bright .mobile-header) {
  --mobile-header-bg: #ffffff;
  --mobile-header-surface: #f3f6f7;
  --mobile-header-border: #dce3e7;
  --mobile-header-text: #23323b;
}

:global(html.bright .mobile-header > .el-button) {
  color: #455761;
  background-color: #f3f6f7;
  border-color: #d6dfe4;
}

:global(html.bright .mobile-header > .el-button:hover),
:global(html.bright .mobile-header > .el-button:focus-visible),
:global(html.bright .mobile-header > .el-button:active) {
  color: #159c9a;
  background-color: #e9f5f5;
  border-color: #9fd6d4;
}

</style>
