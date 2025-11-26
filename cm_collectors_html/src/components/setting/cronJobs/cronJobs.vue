<template>
  <div class="cron-jobs-modern">
    <div class="container">
      <div class="header-section">
        <el-button type="success" plain icon="Plus" @click="createHandle" class="create-btn">
          创建新任务
        </el-button>
      </div>

      <div class="content-wrapper" v-loading="loading">
        <div class="jobs-grid">
          <div v-for="job in cronJobsList" :key="job.id" class="job-card"
            :class="{ 'job-card--failed': !job.last_exec_status && job.last_exec_at }">
            <div class="job-card__header">
              <div class="job-info">
                <span class="job-type" :class="getJobTypeClass(job.jobs_type)">
                  {{ getJobTypeName(job.jobs_type) }}
                </span>
                <h3 class="job-name">{{ store.filesBasesStoreData.getFilesBasesNameById(job.filesBases_id) || '文件库名称' }}
                </h3>
              </div>

              <div class="status-indicator">
                <el-tag :type="getStatusType(job)" size="small" effect="dark">
                  {{ getStatusText(job) }}
                </el-tag>
              </div>
            </div>

            <div class="job-card__body">
              <div class="job-detail">
                <span class="detail-label">Cron表达式</span>
                <span class="detail-value">{{ job.cron_expression }}</span>
              </div>

              <div class="job-detail">
                <span class="detail-label">上次执行</span>
                <span class="detail-value">
                  {{ job.last_exec_at ? formatDate(job.last_exec_at) : '从未执行' }}
                </span>
              </div>

              <div v-if="job.last_exec_error" class="job-detail error-detail">
                <span class="detail-label">错误信息</span>
                <el-tooltip :content="job.last_exec_error" placement="top">
                  <span class="detail-value error-text">
                    {{ truncateText(job.last_exec_error, 40) }}
                  </span>
                </el-tooltip>
              </div>
            </div>

            <div class="job-card__footer">
              <el-button type="primary" link size="small" @click="editJob(job)" class="action-btn">
                <el-icon>
                  <Edit />
                </el-icon>
                编辑
              </el-button>
              <el-button type="danger" link size="small" @click="deleteJob(job)" class="action-btn">
                <el-icon>
                  <Delete />
                </el-icon>
                删除
              </el-button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>

  <createCronJobsDrawer ref="createCronJobsDrawerRef" @success="getCronJobsList"></createCronJobsDrawer>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import createCronJobsDrawer from './createCronJobsDrawer.vue';
import { messageBox, messageBoxConfirm } from '@/common/messageBox';
import { cronJobsServer } from '@/server/cronJobs.server';
import type { I_cronJobs_info } from '@/dataType/cronJobs.dataType';
import { Edit, Delete } from '@element-plus/icons-vue';
import { filesBasesStoreData } from '@/storeData/filesBases.storeData'

const store = {
  filesBasesStoreData: filesBasesStoreData(),
}


const createCronJobsDrawerRef = ref<InstanceType<typeof createCronJobsDrawer>>();
const loading = ref(false);
const cronJobsList = ref<I_cronJobs_info[]>([]);

const init = async () => {
  await getCronJobsList();
}

const getCronJobsList = async () => {
  try {
    loading.value = true;
    const result = await cronJobsServer.list();
    if (result && result.status) {
      cronJobsList.value = result.data.map(item => ({
        ...item,
        filesBasesName: '文件库名称' // 这里应该通过filesBasesId查询实际的文件库名称
      }));
    } else {
      messageBox({ text: result.msg, type: 'error' });
    }
  } catch (error) {
    messageBox({ text: String(error), type: 'error' });
  } finally {
    loading.value = false;
  }
};

const createHandle = () => {
  createCronJobsDrawerRef.value?.open();
};

// 编辑任务
const editJob = (job: I_cronJobs_info) => {
  createCronJobsDrawerRef.value?.open('edit', job);
};

// 删除任务
const deleteJob = (job: I_cronJobs_info) => {
  messageBoxConfirm({
    text: `确定要删除该任务吗？`,
    type: 'warning',
    successCallBack: async () => {
      // 删除任务
      const result = await cronJobsServer.delete(job.id);
      if (result && result.status) {
        messageBox({ text: '删除成功', type: 'success' });
        getCronJobsList();
      } else {
        messageBox({ text: result.msg, type: 'error' });
      }
    },
    failCallBack: () => {
      // 取消
    }
  })
};

// 获取任务类型名称
const getJobTypeName = (type: string): string => {
  switch (type) {
    case 'import': return '导入';
    case 'scraperResource': return '刮削资源';
    case 'scraperPerformer': return '刮削演员';
    case 'clear': return '清理';
    default: return type;
  }
};

// 获取任务类型样式类
const getJobTypeClass = (type: string): string => {
  switch (type) {
    case 'import': return 'job-type--import';
    case 'scraperResource': return 'job-type--scraper';
    case 'scraperPerformer': return 'job-type--performer';
    case 'clear': return 'job-type--clear';
    default: return '';
  }
};

// 获取状态文本
const getStatusText = (job: I_cronJobs_info): string => {
  if (job.last_exec_status) return '成功';
  if (job.last_exec_at) return '失败';
  return '未执行';
};

// 获取状态标签类型
const getStatusType = (job: I_cronJobs_info): 'success' | 'danger' | 'info' => {
  if (job.last_exec_status) return 'success';
  if (job.last_exec_at) return 'danger';
  return 'info';
};

// 格式化日期时间
const formatDate = (dateString: string): string => {
  const date = new Date(dateString);
  return date.toLocaleString('zh-CN');
};

// 截取文本
const truncateText = (text: string, length: number): string => {
  if (text.length > length) {
    return text.substring(0, length) + '...';
  }
  return text;
};

onMounted(() => {
  init();
});
</script>

<style lang="scss" scoped>
.cron-jobs-modern {
  .container {

    .header-section {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 10px 0;

      .create-btn {
        padding: 12px 24px;
        font-size: 14px;
      }
    }

    .content-wrapper {
      min-height: 400px;

      .empty-state {
        background: white;
        border-radius: 8px;
        padding: 60px 20px;
        box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.05);
      }

      .jobs-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
        gap: 20px;

        .job-card {
          background: white;
          border-radius: 6px;
          overflow: hidden;
          box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.05);
          transition: all 0.3s ease;


          &--failed {
            border-left: 4px solid #f56c6c;
          }

          .job-card__header {
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            padding: 10px;
            border-bottom: 1px solid #f0f2f5;

            .job-info {
              .job-type {
                display: inline-block;
                padding: 4px 10px;
                font-size: 12px;
                border-radius: 4px;
                font-weight: 500;
                margin-bottom: 8px;

                &.job-type--import {
                  background-color: #e8f4ff;
                  color: #1890ff;
                }

                &.job-type--scraper {
                  background-color: #fff7e6;
                  color: #fa8c16;
                }

                &.job-type--performer {
                  background-color: #f9f0ff;
                  color: #722ed1;
                }

                &.job-type--clear {
                  background-color: #f6ffed;
                  color: #52c41a;
                }
              }

              .job-name {
                margin: 0;
                font-size: 16px;
                font-weight: 400;
                color: #303133;
              }
            }

            .status-indicator {
              flex-shrink: 0;
            }
          }

          .job-card__body {
            padding: 20px;

            .job-detail {
              display: flex;
              margin-bottom: 16px;

              &:last-child {
                margin-bottom: 0;
              }

              .detail-label {
                width: 90px;
                font-size: 12px;
                color: #606266;
                flex-shrink: 0;
              }

              .detail-value {
                flex: 1;
                font-size: 12px;
                color: #303133;
                word-break: break-all;
              }

              &.error-detail {
                .detail-value {
                  color: #f56c6c;
                }
              }
            }
          }

          .job-card__footer {
            display: flex;
            justify-content: flex-end;
            padding: 16px 20px;
            border-top: 1px solid #f0f2f5;
            gap: 12px;

            .action-btn {
              display: flex;
              align-items: center;
              gap: 4px;
            }
          }
        }
      }
    }
  }
}

@media (max-width: 768px) {
  .cron-jobs-modern {
    padding: 15px;

    .container {
      .header-section {
        flex-direction: column;
        gap: 16px;
        align-items: stretch;
      }

      .content-wrapper {
        .jobs-grid {
          grid-template-columns: 1fr;
        }
      }
    }
  }
}
</style>
