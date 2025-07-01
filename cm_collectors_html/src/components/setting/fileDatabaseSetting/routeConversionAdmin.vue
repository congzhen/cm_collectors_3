<template>
  <div class="routeConversionAdmin">
    <alert-msg color="warning">
      视频文件夹整体移动位置时，可以使用虚拟路径转换功能，例如：from D:\video to E:\myVideo，*!to SoftwareDrive:\myVideo
      则转换至软件当前所在的盘符地址!*，如果需要真实转换数据库中的地址，请使用数据库资源路径替换器。
    </alert-msg>
    <div class="routeAdmin">
      <ul>
        <li v-for="(item, index) in props.routeConversion" :key="index">
          <label>FROM:</label>
          <el-input v-model="item.from" size="small"
            @input="emit('update:routeConversion', props.routeConversion)"></el-input>
          <label>TO:</label>
          <el-input v-model="item.to" size="small"
            @input="emit('update:routeConversion', props.routeConversion)"></el-input>
          <el-button class="coverPosterAdmin-btn" type="danger" icon="delete" size="small"
            @click="deleteRoute(index)"></el-button>
        </li>
      </ul>
    </div>
    <div class="btn">
      <el-button plain @click="addRoute">添加虚拟路径</el-button>
    </div>
  </div>
</template>
<script lang="ts" setup>
import alertMsg from '@/components/com/feedback/alertMsg.vue';
import type { I_routeConversion } from '@/dataType/config.dataType';
const props = defineProps<{
  routeConversion: Array<I_routeConversion>;
}>()

const emit = defineEmits<{
  (e: 'update:routeConversion', val: Array<any>): void;
}>()
const addRoute = () => {
  const newItem = { from: '', to: '' };
  const newData = [...props.routeConversion, newItem];
  emit('update:routeConversion', newData);
}
const deleteRoute = (index: number) => {
  const newData = props.routeConversion.filter((_, i) => i !== index);
  emit('update:routeConversion', newData);
}
</script>
<style lang="scss" scoped>
.routeConversionAdmin {
  .routeAdmin {
    ul {
      list-style-type: none;

      li {
        display: flex;
        padding: 5px 0px;

        label {
          padding: 0 5px;
          line-height: 22px;
          font-size: 12px;
        }

        .el-input {
          width: 200px;
          padding: 0 10px;
        }
      }
    }
  }

  .btn {
    padding-top: 5px;
  }
}
</style>
