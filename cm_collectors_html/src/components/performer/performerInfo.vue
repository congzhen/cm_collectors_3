<template>
  <div class="performer-info">
    <el-alert :title="appLang.performer()" type="success" :closable="false" />
    <div v-if="props.performer">
      <div class="performer-cover">
        <el-image :src="getPerformerPhoto(props.performer)" fit="cover">
          <template #error>
            <el-image src="/emptyPhoto.jpg" fit="cover" />
          </template>
        </el-image>
      </div>
      <div class="performer-info-data">
        <div class="performer-name">{{ props.performer.name }}</div>
        <div><el-rate :model-value="props.performer.stars" disabled /></div>
        <ul class="performer-info-ul">
          <li v-if="props.performer.aliasName != ''">别名：{{ props.performer.aliasName }}</li>
          <li v-if="props.performer.nationality != ''">国籍：{{ props.performer.nationality }}</li>
          <li v-if="store.appStoreData.currentConfigApp.plugInUnit_Cup && props.performer.cup != ''">
            {{ store.appStoreData.currentCupText }}：{{ props.performer.cup }}-{{ store.appStoreData.currentCupText }}
          </li>
          <li v-if="props.performer.bust != '' || props.performer.waist != '' || props.performer.hip != ''">
            <el-breadcrumb>
              <el-breadcrumb-item v-if="props.performer.bust != ''">
                <span>胸围：</span><label>{{ props.performer.bust }}</label>
              </el-breadcrumb-item>
              <el-breadcrumb-item v-if="props.performer.waist != ''">
                <span>腰围：</span><label>{{ props.performer.waist }}</label>
              </el-breadcrumb-item>
              <el-breadcrumb-item v-if="props.performer.hip != ''">
                <span>臀围：</span><label>{{ props.performer.hip }} </label>
              </el-breadcrumb-item>
            </el-breadcrumb>
          </li>
          <li v-if="props.performer.birthday != ''">出生日期：{{ props.performer.birthday }}</li>
          <li v-if="props.performer.introduction != ''">
            简介：{{ props.performer.introduction }}
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>
<script lang="ts" setup>
import type { I_performer } from '@/dataType/performer.dataType';
import { type PropType } from 'vue';
import { appStoreData } from '@/storeData/app.storeData';
import { getPerformerPhoto } from '@/common/photo';
import { appLang } from '@/language/app.lang';
const store = {
  appStoreData: appStoreData(),
}
const props = defineProps({
  performer: {
    type: Object as PropType<I_performer> | undefined,
    default: undefined
  },
});



</script>
<style lang="scss" scoped>
.performer-info {
  .performer-cover {
    width: 100%;
    flex-shrink: 0;
    overflow: hidden;

    .el-image {
      width: 100%;
    }
  }

  .performer-info-data {
    flex-grow: 1;
    padding-left: 0.5em;

    .performer-name {
      font-family: 500;
      font-size: 1.5em;
      color: #ffaa47;
    }

    .performer-info-ul {
      list-style-type: none;
      font-size: 1em;
      line-height: 1.5em;
      color: #cfd3dc;

      :deep(.el-breadcrumb) {
        font-size: 1em;
        line-height: 1.5em;
      }
    }
  }
}
</style>
