<template>
  <div class="performer-details">
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
        <el-button icon="Search" size="small" round @click="showPerforemerResourceHandle"> 查看【{{ props.performer.name
          }}】所有资源 </el-button>
      </div>
    </div>
  </div>
</template>
<script lang="ts" setup>
import type { I_performer } from '@/dataType/performer.dataType';
import { type PropType } from 'vue';
import { calculateAge } from '@/assets/calculate'
import { appStoreData } from '@/storeData/app.storeData';
import { searchStoreData } from '@/storeData/search.storeData';
import { getPerformerPhoto, getPerformerEmptyPhoto } from '@/common/photo';
import performerPhoto from './performerPhoto.vue'
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
  }
})

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
    }
  }
}
</style>
