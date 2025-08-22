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
  router.push('/');
};
</script>

<style lang="scss" scoped>
.mobile-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 15px;
  background-color: #333;
  position: sticky;
  top: 0;
  z-index: 100;

  .title {
    flex: 1;
    text-align: center;
    font-size: 16px;
    font-weight: 500;
    padding: 0 10px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    color: white;
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
      background-color: #333;
      border-radius: 8px;
      width: 200px;

      .menu-item {
        padding: 15px;
        text-align: center;
        border-bottom: 1px solid #444;
        cursor: pointer;
        color: #f3f3f3;

        &:last-child {
          border-bottom: none;
        }

        &:active {
          background-color: #444;
        }
      }
    }
  }
}
</style>
