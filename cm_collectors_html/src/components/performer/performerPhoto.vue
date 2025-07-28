<template>
  <div class="performer-photo">
    <el-image :src="getPerformerPhoto(props.performer)" fit="cover">
      <template #error>
        <el-image src="/emptyPhoto.jpg" fit="cover" />
      </template>
    </el-image>
  </div>
</template>
<script lang="ts" setup>
import type { I_performer } from '@/dataType/performer.dataType';
import { type PropType } from 'vue'
import { getPerformerPhoto } from '@/common/photo';
const props = defineProps({
  performer: {
    type: Object as PropType<I_performer>,
    required: true,
  },
});


</script>
<style lang="scss" scoped>
.performer-photo {
  width: 100%;
  aspect-ratio: 1/1;
  position: relative;
  border-radius: 50%;
  box-shadow: 0 0 0 1px #ffffff;
  overflow: hidden;
  cursor: pointer;

  .el-image {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    border-radius: inherit;

    img {
      border-radius: inherit;
      width: 100%;
      height: 100%;
      object-fit: cover;
    }
  }

  &:hover {
    .el-image {
      transform: scale(1.1);
    }
  }
}
</style>
