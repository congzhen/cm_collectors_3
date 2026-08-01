<template>
  <div class="performer-details" :class="{ 'performer-details--modern': props.modern,
    'performer-details--bright': props.modern && isBrightTheme }">
    <div class="performer-cup" v-if="store.appStoreData.currentConfigApp.plugInUnit_Cup && props.performer.cup != ''">
      {{ props.performer.cup + '-' + store.appStoreData.currentCupText }}
    </div>
    <div class="performer-photo-k">
      <div class="rectangle" v-if="!props.roundAvatar">
        <el-image :src="getPerformerPhoto(props.performer)" fit="cover">
          <template #error>
            <el-image :src="getPerformerEmptyPhoto()" fit="cover" />
          </template>
        </el-image>
      </div>
      <performerPhoto v-else :performer="props.performer"></performerPhoto>
    </div>
    <div class="performer-info">
      <div class="performer-name">{{ props.performer.name }}</div>
      <ul class="performer-info-ul">
        <li v-if="props.performer.aliasName != ''">别名: {{ props.performer.aliasName }}</li>
        <li v-if="props.performer.bust != '' || props.performer.waist != '' || props.performer.hip != ''">
          <el-breadcrumb>
            <el-breadcrumb-item> <span>胸围: </span><label>{{ props.performer.bust }} </label> </el-breadcrumb-item>
            <el-breadcrumb-item> <span>腰围: </span><label>{{ props.performer.waist }} </label> </el-breadcrumb-item>
            <el-breadcrumb-item> <span>臀围: </span><label>{{ props.performer.hip }} </label> </el-breadcrumb-item>
          </el-breadcrumb>
        </li>
        <li v-if="props.performer.birthday != ''">出生日期: {{ props.performer.birthday }}</li>
        <li v-if="props.performer.birthday != '' && props.issuingDate != ''">
          拍摄年龄: {{ calculateAge(props.performer.birthday, props.issuingDate) }}岁
        </li>
        <li>
          {{ props.performer.introduction }}
        </li>
      </ul>
      <div class="performer-btn" v-if="props.performerBtn">
        <el-button icon="Search" size="small" round @click="showPerforemerResourceHandle">
          查看【{{ props.performer.name }}】<label class="res-count" v-loading="resCountStatus">{{ resCount }}</label>部资源
        </el-button>
      </div>
    </div>
  </div>
</template>
<script lang="ts" setup>
import type { I_performer } from '@/dataType/performer.dataType';
import { computed, type PropType, watch, ref } from 'vue';
import { calculateAge } from '@/assets/calculate'
import { appStoreData } from '@/storeData/app.storeData';
import { searchStoreData } from '@/storeData/search.storeData';
import { getPerformerPhoto, getPerformerEmptyPhoto } from '@/common/photo';
import performerPhoto from './performerPhoto.vue'
import { resourceServer } from '@/server/resource.server';
const store = {
  appStoreData: appStoreData(),
  searchStoreData: searchStoreData(),
}
const props = defineProps({
  performer: {
    type: Object as PropType<I_performer>,
    required: true,
  },
  issuingDate: {
    type: String,
    default: ''
  },
  performerBtn: {
    type: Boolean,
    default: true,
  },
  roundAvatar: {
    type: Boolean,
    default: false,
  },
  modern: {
    type: Boolean,
    default: false,
  }
})



const resCount = ref(0)
const resCountStatus = ref(true)
const isBrightTheme = computed(() => store.appStoreData.appConfig.theme === 'bright')

const getPerformerResCount = async (performerId: string) => {
  resCountStatus.value = true
  try {
    const result = await resourceServer.dataCountByPerformerId(store.appStoreData.currentFilesBases.id, performerId);
    if (result && result.status) {
      resCount.value = result.data
    }
  } finally {
    resCountStatus.value = false
  }
}
watch(
  () => props.performer,
  (newPerformer) => {
    getPerformerResCount(newPerformer.id)
    // 在这里可以添加你需要在每次显示时执行的逻辑
  },
  { immediate: true } // 立即触发一次
)

const showPerforemerResourceHandle = () => {
  store.searchStoreData.setQueryPerformer(props.performer.id, props.performer.name)
}


</script>
<style lang="scss" scoped>
.performer-details {
  display: flex;
  position: relative;

  .performer-cup {
    position: absolute;
    z-index: 10;
    right: -6px;
    top: -6px;
    font-weight: bold;
    font-size: 14px;
    color: #F56C6C;
  }

  .performer-photo-k {
    flex-shrink: 0;
    width: 110px;

    .rectangle {
      border-radius: 5px;
      aspect-ratio: 1/1.3;
      overflow: hidden;

      .el-image {
        width: 100%;
        height: 100%;
        border-radius: 5px;
      }
    }

  }

  .performer-info {
    flex-grow: 1;
    padding-left: 15px;
    display: flex;
    flex-direction: column;

    .performer-name {
      flex-shrink: 0;
      font-family: 300;
      font-size: 1.2em;
      color: #ffaa47;
    }

    .performer-info-ul {
      flex-grow: 1;
      list-style-type: none;
      font-size: 0.8em;

      :deep(.el-breadcrumb) {
        .el-breadcrumb__inner {
          font-size: 0.8em;
        }
      }
    }

    .performer-btn {
      flex-shrink: 0;
      transform: scale(0.85);
      display: flex;
      align-items: center;

      .res-count {
        padding: 0 3px;
        font-size: 12px;
        font-weight: 700;
        color: #ffaa47;
      }
    }
  }
}
</style>

<style lang="scss" scoped>
.performer-details--modern {
  --modern-performer-bg: #1f1f1f;
  --modern-performer-soft-bg: #27292d;
  --modern-performer-border: rgba(255, 255, 255, 0.11);
  --modern-performer-text: #e4e7ed;
  --modern-performer-muted: #a8abb2;
  --modern-performer-accent: #37c6ca;

  min-height: 166px;
  padding: 12px;
  box-sizing: border-box;
  color: var(--modern-performer-muted);
  background: var(--modern-performer-bg);
  border-radius: 9px;

  &.performer-details--bright {
    --modern-performer-bg: #ffffff;
    --modern-performer-soft-bg: #f7f8fa;
    --modern-performer-border: #dfe4e9;
    --modern-performer-text: #303a43;
    --modern-performer-muted: #697580;
    --modern-performer-accent: #159fa1;
  }

  .performer-cup {
    top: 12px;
    right: 12px;
    min-width: 54px;
    height: 25px;
    padding: 0 9px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    box-sizing: border-box;
    border: 1px solid var(--modern-performer-accent);
    border-radius: 999px;
    color: var(--modern-performer-accent);
    background: var(--modern-performer-bg);
    font-size: 12px;
    font-weight: 600;
  }

  .performer-photo-k {
    width: 112px;

    .rectangle {
      aspect-ratio: 1 / 1.28;
      border: 1px solid var(--modern-performer-border);
      border-radius: 7px;
      box-sizing: border-box;

      .el-image {
        border-radius: 6px;
      }
    }
  }

  .performer-info {
    min-width: 0;
    padding-left: 14px;

    .performer-name {
      max-width: calc(100% - 82px);
      overflow: hidden;
      color: var(--modern-performer-accent);
      font-size: 18px;
      font-weight: 650;
      line-height: 27px;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .performer-info-ul {
      min-width: 0;
      margin: 4px 0 7px;
      padding: 0;
      color: var(--modern-performer-muted);
      font-size: 12px;
      line-height: 1.65;

      li {
        min-width: 0;
        overflow-wrap: anywhere;
      }

      li:last-child {
        display: -webkit-box;
        overflow: hidden;
        -webkit-box-orient: vertical;
        -webkit-line-clamp: 2;
      }

      :deep(.el-breadcrumb) {
        line-height: inherit;

        .el-breadcrumb__inner,
        .el-breadcrumb__separator {
          color: var(--modern-performer-muted);
          font-size: 12px;
          font-weight: 400;
        }
      }
    }

    .performer-btn {
      width: fit-content;
      max-width: 100%;
      margin-top: auto;
      align-self: flex-start;
      transform: none;

      :deep(.el-button) {
        width: auto;
        max-width: 100%;
        height: 28px;
        padding: 0 13px;
        justify-content: center;
        color: var(--modern-performer-text);
        background: var(--modern-performer-soft-bg);
        border-color: var(--modern-performer-border);
        border-radius: 6px;

        &:hover {
          color: var(--modern-performer-accent);
          border-color: var(--modern-performer-accent);
        }
      }

      .res-count {
        color: #e89432;
      }
    }
  }
}
</style>
