<template>
  <div class="header-container">
    <div class="title">
      <img src="/public/icon32.png" />
      <label>{{ store.appStoreData.getLogoName }}</label>
      <switchMobile></switchMobile>
    </div>
    <div class="right" v-if="props.mode === E_headerMode.Index">
      <div class="search">
        <searchInputTagByStore />
      </div>
      <div class="setting">
        <div class="menu-item" v-admin @click="oepnResourceForm">
          <el-tooltip content="添加" placement="bottom">
            <el-icon>
              <Plus />
            </el-icon>
          </el-tooltip>
          <!--
           <label class="menu-item-title">添加</label>
          -->
        </div>
        <div class="menu-item" v-admin @click="switchAdminStatus">
          <el-tooltip :content="!store.appStoreData.adminResourceStatus ? '管理关闭' : '管理开启'" placement="bottom">
            <el-icon v-if="!store.appStoreData.adminResourceStatus">
              <TurnOff />
            </el-icon>
            <el-icon v-else color="#409EFF">
              <Open />
            </el-icon>
          </el-tooltip>
          <!--
          <label class="menu-item-title">管理</label>
           -->
        </div>
        <div class="menu-item" v-admin @click="openTagList">
          <el-tooltip content="标签" placement="bottom">
            <el-icon>
              <PriceTag />
            </el-icon>
          </el-tooltip>
          <!--
          <label class="menu-item-title">标签</label>
           -->
        </div>
        <div class="menu-item" @click="goToPerformer">
          <el-tooltip :content="appLang.performer()" placement="bottom">
            <el-icon>
              <User />
            </el-icon>
          </el-tooltip>
          <!--
          <label class="menu-item-title">{{ appLang.performer() }}</label>
           -->
        </div>
        <div class="menu-item" v-admin @click="openImportResource">
          <el-tooltip content="导入与刮削" placement="bottom">
            <el-icon>
              <Magnet />
            </el-icon>
          </el-tooltip>
          <!--
          <label class="menu-item-title">导入</label>
           -->
        </div>
        <div class="menu-item" v-admin @click="openCheckUpdateSoft">
          <el-tooltip content="检测更新" placement="bottom">
            <el-icon>
              <ChromeFilled />
            </el-icon>
          </el-tooltip>
          <!--
          <label class="menu-item-title">设置</label>
           -->
        </div>
        <div class="menu-item" v-admin @click="goToSetting">
          <el-tooltip content="设置" placement="bottom">
            <el-icon>
              <Setting />
            </el-icon>
          </el-tooltip>
          <!--
          <label class="menu-item-title">设置</label>
           -->
        </div>
        <div class="menu-item" v-if="store.appStoreData.isAdminLogin && !store.appStoreData.isAdminLoginStatus">
          <el-link type="primary" href="/adminLogin">登录</el-link>
        </div>
      </div>
    </div>
    <div class="right" v-else>
      <div class="setting sub-nav">
        <el-tooltip content="主页" placement="bottom">
          <label class="icon-text-label" @click="router.push('/')">
            <el-icon title="主页">
              <HomeFilled />
            </el-icon>
          </label>
        </el-tooltip>
        <el-tooltip content="播放列表" placement="bottom">
          <label class="icon-text-label" @click="openPlayListHandle">
            <el-icon title="播放列表">
              <Memo />
            </el-icon>
          </label>
        </el-tooltip>
        <el-tooltip content="返回" placement="bottom">
          <label class="icon-text-label" @click="router.go(-1)">
            <el-icon title="返回">
              <Back />
            </el-icon>
          </label>
        </el-tooltip>
      </div>
    </div>
    <tagListDrawer ref="tagListDrawerRef" />
    <resourceFormDrawer ref="resourceFormDrawerRef" @success="createResouceSuccessHandle" />
    <importResourceDrawer ref="importResourceDrawerRef" @success="createResouceSuccessHandle" />
    <updateSoftDialog ref="updateSoftDialogRef"></updateSoftDialog>
    <playListDrawer ref="playListDrawerRef"></playListDrawer>
  </div>
</template>
<script setup lang="ts">
import { ref, type PropType } from 'vue'
import { E_headerMode } from '@/dataType/app.dataType'
import { useRouter } from 'vue-router'
import switchMobile from '@/components/com/form/switchMobile.vue';
import tagListDrawer from '@/components/tag/tagListDrawer.vue'
import searchInputTagByStore from '@/components/com/form/searchInputTagByStore.vue'
import resourceFormDrawer from '@/components/resource/resourceFormDrawer.vue'
import importResourceDrawer from '@/components/importResource/importResourceDrawer.vue'
import updateSoftDialog from '@/components/setting/updateSoft/updateSoftDialog.vue'
import playListDrawer from '@/components/playList/playListDrawer.vue'
import { appStoreData } from '@/storeData/app.storeData'
import type { I_resource } from '@/dataType/resource.dataType'
import { AppLang } from '@/language/app.lang'
const appLang = AppLang()

const router = useRouter()
const store = {
  appStoreData: appStoreData(),
}
const emits = defineEmits(['createResouceSuccess'])

const tagListDrawerRef = ref<InstanceType<typeof tagListDrawer>>()
const resourceFormDrawerRef = ref<InstanceType<typeof resourceFormDrawer>>()
const importResourceDrawerRef = ref<InstanceType<typeof importResourceDrawer>>()
const updateSoftDialogRef = ref<InstanceType<typeof updateSoftDialog>>()
const playListDrawerRef = ref<InstanceType<typeof playListDrawer>>()

const props = defineProps({
  mode: {
    type: String as PropType<E_headerMode>,
    default: E_headerMode.Index,
  },
})


const oepnResourceForm = () => {
  resourceFormDrawerRef.value?.open('add')
}

const switchAdminStatus = () => {
  store.appStoreData.adminResourceStatus = !store.appStoreData.adminResourceStatus
}

const openTagList = () => {
  tagListDrawerRef.value?.open()
}

const openImportResource = () => {
  importResourceDrawerRef.value?.open()
}

const openCheckUpdateSoft = () => {
  updateSoftDialogRef.value?.open()
}
const openPlayListHandle = () => {
  playListDrawerRef.value?.open()
}

const goToPerformer = () => {
  if (store.appStoreData.currentPerformerBasesIds.length > 1) {
    router.push(`/performer/basesList/${store.appStoreData.currentFilesBases.id}`)
  } else {
    router.push(`/performer/${store.appStoreData.currentMainPerformerBasesId}`)
  }
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
    padding: 0;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
    display: flex;
    align-items: center;
    gap: 5px;


    img {
      width: 1em;
      height: 1em;
    }
  }

  .right {
    display: flex;
    justify-content: flex-end;
    padding-right: 0.5em;
    flex-grow: 1;

    .search {
      margin: 0 1em;
      width: 70%;
    }


    .setting {
      flex-shrink: 0;
      font-size: 2em;
      cursor: pointer;
      display: flex;
      justify-content: flex-end;

      .menu-item {
        margin-left: 0.2em;
        display: flex;
        align-items: center;
        cursor: pointer;
        user-select: none;

        &:hover {
          color: var(--el-color-primary);

          .menu-item-title {
            display: block;
            width: auto;
            opacity: 1;
          }
        }

        .menu-item-title {
          font-size: 14px;
          line-height: 28px;
          cursor: pointer;
          width: 0;
          opacity: 0;
          transition: width 0.3s ease, opacity 0.1s ease;
          white-space: nowrap;
          overflow: hidden;
        }

        .el-link {
          padding: 2px 5px;
        }
      }

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
    }

    .sub-nav {
      display: flex;
      align-items: center;
      gap: 5px;

      .el-icon {
        font-size: 20px;
      }
    }
  }
}
</style>
