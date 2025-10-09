<template>
  <div class="coverPosterAdmin">
    <alertMsg color="warning">如需遵照尺寸显示封面，需关闭宽高控制</alertMsg>
    <ul>
      <el-radio-group v-model="props.coverPosterDataDefaultSelect"
        @change="emit('update:coverPosterDataDefaultSelect', $event)">
        <li v-for="(item, index) in coverPosterData" :key="index">
          <el-radio class="coverPosterAdmin-radio" :value="index">
            <div class="coverPosterAdmin-radio-block">
              <div class="coverPosterAdmin-name">
                <el-input v-if="item.type == 'default'" v-model="item.name" size="small" :disabled="true"></el-input>
                <el-input v-else v-model="item.name" @input="emit('update:coverPosterData', props.coverPosterData)"
                  size="small"></el-input>
              </div>
              <label class="coverPosterAdmin-text">宽:</label>
              <el-input-number v-model="item.width" @input="emit('update:coverPosterData', props.coverPosterData)"
                controls-position="right" size="small" :min="20" :max="1280"
                :disabled="item.type == 'default' ? true : false"></el-input-number>
              <label class="coverPosterAdmin-text">高:</label>
              <el-input-number v-model="item.height" @input="emit('update:coverPosterData', props.coverPosterData)"
                controls-position="right" size="small" :min="20" :max="800"
                :disabled="item.type == 'default' ? true : false"></el-input-number>
              <label class="coverPosterAdmin-bz">
                {{ coverPosterBz(item.width, item.height) }}
              </label>
              <el-button class="coverPosterAdmin-btn" type="danger" icon="delete" size="small"
                v-if="item.type != 'default'" @click="deleteCoverPoster(index)"></el-button>
            </div>

          </el-radio>
        </li>
      </el-radio-group>
    </ul>
    <div>
      <el-button plain @click="addCoverPoster">添加新封面海报尺寸</el-button>
    </div>
  </div>
</template>
<script lang="ts" setup>
import alertMsg from '@/components/com/feedback/alertMsg.vue';
import { ratio } from '@/assets/calculate'
import type { I_coverPosterData } from '@/dataType/config.dataType';
const props = defineProps<{
  coverPosterDataDefaultSelect: number;
  coverPosterData: Array<I_coverPosterData>;
}>()

const emit = defineEmits<{
  (e: 'update:coverPosterDataDefaultSelect', val: number): void;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  (e: 'update:coverPosterData', val: Array<any>): void;
}>()
const coverPosterBz = (width: number, height: number) => {
  return ratio(width, height).join(' : ');
}

const addCoverPoster = () => {
  const newItem: I_coverPosterData = { name: 'DiyCover', width: 480, height: 270, type: 'diy' };
  const newData = [...props.coverPosterData, newItem];
  emit('update:coverPosterData', newData);
}
const deleteCoverPoster = (index: number) => {
  const newData = props.coverPosterData.filter((_, i) => i !== index);
  emit('update:coverPosterData', newData);
}
</script>
<style lang="scss" scoped>
.coverPosterAdmin {
  ul {
    list-style-type: none;
    padding: 10px 20px 10px 20px;

    li {
      display: flex;
      padding: 5px 0px;
      line-height: 28px;

      .coverPosterAdmin-radio {
        height: 28px;

      }

      .coverPosterAdmin-name {
        width: 120px;
      }

      .coverPosterAdmin-radio-block {
        display: flex;
      }

      .coverPosterAdmin-text {
        padding: 0px 10px 0px 20px;
      }

      .coverPosterAdmin-bz {
        min-width: 100px;
        padding-left: 20px;
        overflow: hidden;
      }

      .coverPosterAdmin-btn {
        margin-top: 2px;
      }
    }
  }
}
</style>
