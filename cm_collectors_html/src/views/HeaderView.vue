<template>
  <div class="header-container">
    <div class="title">
      <img src="/public/icon32.png" />
      <label>{{ props.title }}</label>
    </div>
    <div class="right" v-if="props.mode === E_headerMode.Index">
      <div class="search">
        <searchInputTagByStore />
      </div>
      <div class="setting">
        <el-icon title="添加" @click="oepnResourceForm">
          <Plus />
        </el-icon>
        <el-icon v-if="!store.appStoreData.adminStatus" @click="switchAdminStatus">
          <TurnOff />
        </el-icon>
        <el-icon v-else color="#409EFF" @click="switchAdminStatus">
          <Open />
        </el-icon>
        <el-icon title="标签" @click="openTagList">
          <PriceTag />
        </el-icon>
        <el-icon title="演员" @click="goToPerformer">
          <User />
        </el-icon>
        <el-icon title="设置" @click="goToSetting">
          <Setting />
        </el-icon>
      </div>
    </div>
    <div class="right" v-else>
      <div class="setting">
        <label class="icon-text-label" @click="router.go(-1)">
          <el-icon title="返回">
            <Back />
          </el-icon>
          <span class="icon-text-span">返回</span>
        </label>
      </div>
    </div>
    <tagListDrawer ref="tagListDrawerRef" />
    <resourceFormDrawer ref="resourceFormDrawerRef" @success="createResouceSuccessHandle" />
  </div>
</template>
<script setup lang="ts">
import { ref, type PropType } from 'vue'
import { E_headerMode } from '@/dataType/app.dataType'
import { useRouter } from 'vue-router'
import tagListDrawer from '@/components/tag/tagListDrawer.vue'
import searchInputTagByStore from '@/components/com/form/searchInputTagByStore.vue'
import resourceFormDrawer from '@/components/resource/resourceFormDrawer.vue'
import { appStoreData } from '@/storeData/app.storeData'
import type { I_resource } from '@/dataType/resource.dataType'
const router = useRouter()
const store = {
  appStoreData: appStoreData(),
}
const emits = defineEmits(['createResouceSuccess'])

const tagListDrawerRef = ref<InstanceType<typeof tagListDrawer>>()
const resourceFormDrawerRef = ref<InstanceType<typeof resourceFormDrawer>>()

const props = defineProps({
  title: {
    type: String,
    default: 'CM File Collectors',
  },
  mode: {
    type: String as PropType<E_headerMode>,
    default: E_headerMode.Index,
  },
})


const oepnResourceForm = () => {
  resourceFormDrawerRef.value?.open('add')
}

const switchAdminStatus = () => {
  store.appStoreData.adminStatus = !store.appStoreData.adminStatus
}

const openTagList = () => {
  tagListDrawerRef.value?.open()
}
const goToPerformer = () => {
  router.push(`/performer/${store.appStoreData.currentMainPerformerBasesId}`)
}
const goToSetting = () => {
  router.push('/setting')
}
const createResouceSuccessHandle = (data: I_resource) => {
  emits('createResouceSuccess', data)
}
</script>
<style lang="scss" scoped>
.header-container {
  display: flex;
  justify-content: space-between;


  .title {
    font-size: 1.5em;
    font-weight: 500;
    flex-shrink: 0;
    max-width: 50%;
    padding: 0;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;


    img {
      width: 1em;
      height: 1em;
      margin-right: 0.3em;
    }
  }

  .right {
    display: flex;
    justify-content: flex-end;
    padding-right: 0.5em;
    flex-grow: 1;

    .search {
      margin: 0 1em;
      width: 75%;
    }


    .setting {
      flex-shrink: 0;
      font-size: 2em;
      cursor: pointer;

      .icon-text-label {
        display: flex;
        cursor: pointer;

        &:hover {
          color: var(--el-color-primary);
        }

        .icon-text-span {
          font-size: 0.65em;
        }
      }

      .el-icon {
        margin-left: 0.2em;

        &:hover {
          color: var(--el-color-primary);
        }
      }
    }
  }
}
</style>
