<template>
  <div class="create-cron-jobs">
    <drawerCommon ref="drawerCommonRef" width="680px" title="计划任务" @submit="submitHandle">
      <div class="create-cron-jobs-main">
        <el-form ref="ruleFormRef" :model="formData" label-width="160px" label-position="top" status-icon>
          <el-form-item label="执行文件库">
            <selectFilesBases v-model="formData.filesBasesIds" multiple></selectFilesBases>
          </el-form-item>
          <el-form-item label="任务类型">
            <el-radio-group v-model="formData.jobsType">
              <el-radio-button label="导入" value="import" />
              <el-radio-button label="刮削资源" value="scraperResource" />
              <el-radio-button label="刮削演员" value="scraperPerformer" />
              <el-radio-button label="清理" value="clear" />
            </el-radio-group>
          </el-form-item>

          <el-form-item label="计划任务设定 (Cron表达式)">
            <div class="cron-expression-container">
              <div class="cron-expression-row">
                <el-select v-model="cronParts.second" placeholder="秒">
                  <el-option v-for="item in secondsOptions" :key="item.value" :label="item.label" :value="item.value" />
                </el-select>
                <el-select v-model="cronParts.minute" placeholder="分">
                  <el-option v-for="item in minutesOptions" :key="item.value" :label="item.label" :value="item.value" />
                </el-select>
                <el-select v-model="cronParts.hour" placeholder="时">
                  <el-option v-for="item in hoursOptions" :key="item.value" :label="item.label" :value="item.value" />
                </el-select>
                <el-select v-model="cronParts.day" placeholder="日">
                  <el-option v-for="item in daysOptions" :key="item.value" :label="item.label" :value="item.value" />
                </el-select>
                <el-select v-model="cronParts.month" placeholder="月">
                  <el-option v-for="item in monthsOptions" :key="item.value" :label="item.label" :value="item.value" />
                </el-select>
                <el-select v-model="cronParts.week" placeholder="周">
                  <el-option v-for="item in weeksOptions" :key="item.value" :label="item.label" :value="item.value" />
                </el-select>
              </div>
              <div class="cron-help-text">
                秒 分 时 日 月 周
              </div>
            </div>
            <div class="cron-preview-container">
              <div class="cron-preview">
                <div class="preview-title">表达式预览:</div>
                <div class="preview-content">{{ generateCronExpression() }}</div>
              </div>
              <div class="cron-description">
                <div class="preview-title">通俗描述:</div>
                <div class="description-content">{{ generateHumanReadableDescription() }}</div>
              </div>
            </div>
          </el-form-item>
        </el-form>
      </div>
    </drawerCommon>
  </div>
</template>
<script setup lang="ts">
import drawerCommon from '@/components/com/dialog/drawer-common.vue';
import selectFilesBases from '@/components/com/form/selectFilesBases.vue';
import { ref, reactive } from 'vue';

// 定义各个字段的选项
const secondsOptions = [
  { label: '0', value: '0' },
  { label: '5', value: '5' },
  { label: '10', value: '10' },
  { label: '15', value: '15' },
  { label: '20', value: '20' },
  { label: '25', value: '25' },
  { label: '30', value: '30' },
  { label: '35', value: '35' },
  { label: '40', value: '40' },
  { label: '45', value: '45' },
  { label: '50', value: '50' },
  { label: '55', value: '55' },
  { label: '每5秒', value: '*/5' },
  { label: '每10秒', value: '*/10' },
  { label: '每30秒', value: '*/30' },
  { label: '每秒', value: '*' }
];

const minutesOptions = [
  { label: '0', value: '0' },
  { label: '5', value: '5' },
  { label: '10', value: '10' },
  { label: '15', value: '15' },
  { label: '20', value: '20' },
  { label: '25', value: '25' },
  { label: '30', value: '30' },
  { label: '35', value: '35' },
  { label: '40', value: '40' },
  { label: '45', value: '45' },
  { label: '50', value: '50' },
  { label: '55', value: '55' },
  { label: '每5分钟', value: '*/5' },
  { label: '每10分钟', value: '*/10' },
  { label: '每30分钟', value: '*/30' },
  { label: '每分钟', value: '*' }
];

const hoursOptions = [
  { label: '0', value: '0' },
  { label: '1', value: '1' },
  { label: '2', value: '2' },
  { label: '3', value: '3' },
  { label: '4', value: '4' },
  { label: '5', value: '5' },
  { label: '6', value: '6' },
  { label: '7', value: '7' },
  { label: '8', value: '8' },
  { label: '9', value: '9' },
  { label: '10', value: '10' },
  { label: '11', value: '11' },
  { label: '12', value: '12' },
  { label: '13', value: '13' },
  { label: '14', value: '14' },
  { label: '15', value: '15' },
  { label: '16', value: '16' },
  { label: '17', value: '17' },
  { label: '18', value: '18' },
  { label: '19', value: '19' },
  { label: '20', value: '20' },
  { label: '21', value: '21' },
  { label: '22', value: '22' },
  { label: '23', value: '23' },
  { label: '每小时', value: '*' }
];

const daysOptions = [
  { label: '1', value: '1' },
  { label: '2', value: '2' },
  { label: '3', value: '3' },
  { label: '4', value: '4' },
  { label: '5', value: '5' },
  { label: '6', value: '6' },
  { label: '7', value: '7' },
  { label: '8', value: '8' },
  { label: '9', value: '9' },
  { label: '10', value: '10' },
  { label: '11', value: '11' },
  { label: '12', value: '12' },
  { label: '13', value: '13' },
  { label: '14', value: '14' },
  { label: '15', value: '15' },
  { label: '16', value: '16' },
  { label: '17', value: '17' },
  { label: '18', value: '18' },
  { label: '19', value: '19' },
  { label: '20', value: '20' },
  { label: '21', value: '21' },
  { label: '22', value: '22' },
  { label: '23', value: '23' },
  { label: '24', value: '24' },
  { label: '25', value: '25' },
  { label: '26', value: '26' },
  { label: '27', value: '27' },
  { label: '28', value: '28' },
  { label: '29', value: '29' },
  { label: '30', value: '30' },
  { label: '31', value: '31' },
  { label: '每日', value: '*' }
];

const monthsOptions = [
  { label: '1', value: '1' },
  { label: '2', value: '2' },
  { label: '3', value: '3' },
  { label: '4', value: '4' },
  { label: '5', value: '5' },
  { label: '6', value: '6' },
  { label: '7', value: '7' },
  { label: '8', value: '8' },
  { label: '9', value: '9' },
  { label: '10', value: '10' },
  { label: '11', value: '11' },
  { label: '12', value: '12' },
  { label: '每月', value: '*' }
];

const weeksOptions = [
  { label: '0(日)', value: '0' },
  { label: '1(一)', value: '1' },
  { label: '2(二)', value: '2' },
  { label: '3(三)', value: '3' },
  { label: '4(四)', value: '4' },
  { label: '5(五)', value: '5' },
  { label: '6(六)', value: '6' },
  { label: '7(日)', value: '7' },
  { label: '每日', value: '*' }
];

const drawerCommonRef = ref<InstanceType<typeof drawerCommon>>();
const formData = ref({
  jobsType: 'import',
  filesBasesIds: [],
  cronExpression: ''
})

const cronParts = reactive({
  second: '0',
  minute: '0',
  hour: '2',
  day: '*',
  month: '*',
  week: '*'
});

// 生成完整的 cron 表达式
const generateCronExpression = () => {
  return `${cronParts.second} ${cronParts.minute} ${cronParts.hour} ${cronParts.day} ${cronParts.month} ${cronParts.week}`;
};

// 生成人类可读的描述
const generateHumanReadableDescription = () => {
  let desc = "在";

  // 处理月份
  if (cronParts.month === '*') {
    desc += "每个月的";
  } else {
    desc += `${cronParts.month}月的`;
  }

  // 处理日期和星期
  if (cronParts.day !== '*' && cronParts.week !== '*') {
    desc += `${cronParts.day}日(忽略星期)`;
  } else if (cronParts.day !== '*') {
    desc += `${cronParts.day}日`;
  } else if (cronParts.week !== '*') {
    const weekdays = ['日', '一', '二', '三', '四', '五', '六'];
    if (parseInt(cronParts.week) <= 7) {
      desc += `每周${weekdays[parseInt(cronParts.week) % 7]}`;
    } else {
      desc += "每天";
    }
  } else {
    desc += "每天";
  }

  // 处理小时
  if (cronParts.hour === '*') {
    desc += "每小时";
  } else {
    desc += `的${cronParts.hour}点`;
  }

  // 处理分钟
  if (cronParts.minute === '*') {
    if (cronParts.hour !== '*') {
      desc += "的每分钟";
    }
  } else {
    desc += `${cronParts.minute}分`;
  }

  // 处理秒钟
  if (cronParts.second === '*') {
    if (cronParts.minute !== '*' || cronParts.hour !== '*') {
      desc += "的每一秒";
    }
  } else {
    desc += `${cronParts.second}秒`;
  }

  desc += "执行";
  return desc;
};

// 监听 cronParts 的变化并更新 formData.cronExpression
import { watch } from 'vue';
watch(cronParts, () => {
  formData.value.cronExpression = generateCronExpression();
}, { deep: true });

const init = () => {
  // 初始化 cronParts
  cronParts.second = '0';
  cronParts.minute = '0';
  cronParts.hour = '2';
  cronParts.day = '*';
  cronParts.month = '*';
  cronParts.week = '*';

  // 初始化 formData.cronExpression
  formData.value.cronExpression = generateCronExpression();
}

const submitHandle = () => {

}

const open = async () => {
  drawerCommonRef.value?.open();
  init();
}

const colse = () => {
  drawerCommonRef.value?.close();
}

defineExpose({ open })
</script>
<style lang="scss" scoped>
.cron-expression-row {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;

  .el-select {
    flex: 1;
    min-width: 80px;
  }
}

.cron-help-text {
  font-size: 12px;
  color: #999;
  margin-top: 5px;
}

.cron-expression-container {
  width: 100%;
  display: flex;
  justify-content: space-between;
}

.cron-preview-container {
  width: 100%;
  margin-top: 15px;
  border-radius: 6px;
}

.preview-title {
  font-weight: bold;
  margin-bottom: 5px;
  font-size: 14px;
}

.preview-content {
  font-family: 'Courier New', monospace;
  background-color: #212529;
  padding: 8px 12px;
  border-radius: 4px;
  font-size: 14px;
  word-break: break-all;
  line-height: 36px;
}

.cron-description {
  margin-top: 15px;
}

.description-content {
  background-color: #212529;
  padding: 8px 12px;
  border-radius: 4px;
  font-size: 14px;
  line-height: 36px;
}
</style>
