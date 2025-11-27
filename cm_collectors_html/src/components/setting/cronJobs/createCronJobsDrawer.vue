<template>
  <div class="create-cron-jobs">
    <drawerCommon ref="drawerCommonRef" width="720px" title="计划任务" @submit="submitHandle">
      <div class="create-cron-jobs-main" :loading="loading">
        <el-form ref="ruleFormRef" :model="formData" label-width="160px" label-position="top" status-icon>
          <el-form-item label="执行文件库">
            <selectFilesBases v-model="formData.filesBases_id"></selectFilesBases>
          </el-form-item>
          <el-form-item label="任务类型">
            <el-radio-group v-model="formData.jobs_type">
              <el-radio-button label="导入" value="import" />
              <el-radio-button label="刮削资源" value="scraperResource" />
              <el-radio-button label="刮削演员" value="scraperPerformer" />
              <el-radio-button label="清理" value="clear" />
            </el-radio-group>
            <el-text class="warning-text" type="warning" size="small">
              每种任务类型都依赖于预先设定的功能配置，缺少相应配置将导致任务无法执行。
            </el-text>
          </el-form-item>
          <el-form-item label="计划任务设定 (Cron表达式: 秒 分 时 日 月 周)">
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
            </div>
            <div class="cron-preview-container">
              <div class="cron-preview">
                <div class="preview-title">表达式预览:</div>
                <div class="preview-content">{{ generatecron_expression() }}</div>
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
import { debounceNow } from '@/assets/debounce';
import { messageBoxAlert } from '@/common/messageBox';
import drawerCommon from '@/components/com/dialog/drawer-common.vue';
import selectFilesBases from '@/components/com/form/selectFilesBases.vue';
import type { I_cronJobs, I_cronJobs_info } from '@/dataType/cronJobs.dataType';
import { cronJobsServer } from '@/server/cronJobs.server';
import { ref, reactive, nextTick } from 'vue';

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

const emits = defineEmits(['success'])

const drawerCommonRef = ref<InstanceType<typeof drawerCommon>>();
const formData = ref<I_cronJobs>({
  id: '',
  jobs_type: 'import',
  filesBases_id: '',
  cron_expression: ''
})
const loading = ref(false);
let mode: 'add' | 'edit' = 'add';

const cronParts = reactive({
  second: '0',
  minute: '0',
  hour: '2',
  day: '*',
  month: '*',
  week: '*'
});

// 生成完整的 cron 表达式
const generatecron_expression = () => {
  return `${cronParts.second} ${cronParts.minute} ${cronParts.hour} ${cronParts.day} ${cronParts.month} ${cronParts.week}`;
};

// 生成人类可读的描述
const generateHumanReadableDescription = () => {
  let desc = "在";

  // 处理月份
  if (cronParts.month === '*') {
    desc += "每个月";
  } else {
    desc += `${cronParts.month}月`;
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
    desc += `${cronParts.hour}点`;
  }

  // 处理分钟
  if (cronParts.minute === '*') {
    //if (cronParts.hour !== '*') {
    desc += "每分钟";
    //}
  } else {
    desc += `${cronParts.minute}分`;
  }

  // 处理秒钟
  if (cronParts.second === '*') {
    if (cronParts.minute !== '*' || cronParts.hour !== '*') {
      desc += "每一秒";
    }
  } else {
    desc += `${cronParts.second}秒`;
  }

  desc += "执行";
  return desc;
};

// 监听 cronParts 的变化并更新 formData.cron_expression
import { watch } from 'vue';
watch(cronParts, () => {
  formData.value.cron_expression = generatecron_expression();
}, { deep: true });

const init = (_mode: 'add' | 'edit' = 'add', _info: I_cronJobs_info | null = null) => {
  mode = _mode;
  if (_mode == 'edit' && _info != null) {
    formData.value.id = _info.id;
    formData.value.filesBases_id = _info.filesBases_id;
    formData.value.jobs_type = _info.jobs_type;

    // 将传入的 cron 表达式拆分成 cronParts
    const parts = _info.cron_expression.split(' ');
    if (parts.length === 6) {
      cronParts.second = parts[0];
      cronParts.minute = parts[1];
      cronParts.hour = parts[2];
      cronParts.day = parts[3];
      cronParts.month = parts[4];
      cronParts.week = parts[5];
    } else {
      // 如果表达式格式不正确，使用默认值
      cronParts.second = '0';
      cronParts.minute = '0';
      cronParts.hour = '2';
      cronParts.day = '*';
      cronParts.month = '*';
      cronParts.week = '*';
    }
  } else {
    formData.value.id = '';
    formData.value.filesBases_id = '';
    formData.value.jobs_type = 'import';
    // 初始化 cronParts
    cronParts.second = '0';
    cronParts.minute = '0';
    cronParts.hour = '2';
    cronParts.day = '*';
    cronParts.month = '*';
    cronParts.week = '*';
  }



  // 初始化 formData.cron_expression
  formData.value.cron_expression = generatecron_expression();
}

const submitHandle = debounceNow(async () => {
  console.log(formData.value)
  if (formData.value.filesBases_id === '') {
    messageBoxAlert({ text: '请选择执行文件库', type: 'warning' });
    return;
  }
  try {
    loading.value = true;
    let result;
    let successText: string;
    if (mode == 'add') {
      result = await cronJobsServer.create(formData.value.filesBases_id, formData.value.jobs_type, formData.value.cron_expression)
      successText = '创建成功'
    } else {
      result = await cronJobsServer.update(formData.value.id, formData.value.filesBases_id, formData.value.jobs_type, formData.value.cron_expression)
      successText = '修改成功'
    }
    if (result && result.status) {
      messageBoxAlert({ text: successText, type: 'success' });
      emits('success')
      colse();
    } else {
      messageBoxAlert({ text: result.msg, type: 'error' });
    }
  } catch (error) {
    messageBoxAlert({ text: String(error), type: 'error' });
  } finally {
    loading.value = false;
  }
})

const open = async (_mode: 'add' | 'edit' = 'add', _info: I_cronJobs_info | null = null) => {
  init(_mode, _info);
  nextTick(() => {
    drawerCommonRef.value?.open();
  })
}

const colse = () => {
  drawerCommonRef.value?.close();
}

defineExpose({ open })
</script>
<style lang="scss" scoped>
.cron-expression-row {
  display: flex;
  gap: 5px;
  flex-wrap: wrap;

  .el-select {
    flex: 1;
    min-width: 105px;
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
