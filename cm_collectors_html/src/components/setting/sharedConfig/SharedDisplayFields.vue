<template>
  <div class="shared-fields">
    <p v-if="readonly" class="shared-source-hint">通用参数来自公共配置，下方只读展示；本库独立项仍可编辑。修改通用参数请点击上方“打开公共配置”。</p>
    <SettingSectionTitle>基础设置</SettingSectionTitle>
    <SettingFieldsScope :disabled="readonly">
      <el-form-item label="国家列表">
        <selectCountry v-model="config.country" multiple />
      </el-form-item>
      <el-form-item label="清晰度">
        <selectDefinition v-model="config.definition" multiple />
      </el-form-item>
      <el-form-item label="资源排序">
        <selectResourceSort v-model="config.resourceSort" multiple />
      </el-form-item>
    </SettingFieldsScope>
    <SettingSectionTitle>左侧边栏</SettingSectionTitle>
    <SettingFieldsScope :disabled="readonly">
      <el-form-item label="左侧边栏显示项">
        <selectLeftDisplay v-model="config.leftDisplay" multiple />
      </el-form-item>
      <el-form-item label="左侧边栏显示模式">
        <selectLeftColumnMode v-model="config.leftColumnMode" />
      </el-form-item>
      <el-form-item label="左侧边栏宽度">
        <el-input-number v-model="config.leftColumnWidth" :min="100" />
      </el-form-item>
      <el-form-item label="浮动模式自动隐藏">
        <el-checkbox
          v-model="config.leftColumnFloatAutoHide"
          label="点击非左侧栏位置后自动隐藏"
          border
        />
      </el-form-item>
      <el-form-item label="标签显示模式">
        <selectTagMode v-model="config.tagMode" />
      </el-form-item>
      <el-form-item label="固定模式每行显示标签数量">
        <el-input-number v-model="config.tagFixedModeRowShowNum" :min="1" :max="99" />
      </el-form-item>
      <el-form-item label="自定义标签资源数量">
        <el-switch
          v-model="config.showCustomTagResourceCount"
          inline-prompt
          active-text="显示"
          inactive-text="隐藏"
        />
      </el-form-item>
      <el-form-item label="演员标签">
        <el-checkbox v-model="config.performerPhoto" label="显示演员照片" border />
        <el-checkbox v-model="config.shieldNoPerformerPhoto" label="屏蔽无照片演员" border />
      </el-form-item>
      <el-form-item label="演员标签显示数量">
        <el-input-number v-model="config.performerShowNum" />
      </el-form-item>
    </SettingFieldsScope>
    <slot name="actors" />
    <SettingSectionTitle>显示设置</SettingSectionTitle>
    <SettingFieldsScope :disabled="readonly">
      <el-form-item label="分页显示数量">
        <el-input-number v-model="config.pageLimit" />
      </el-form-item>
      <el-form-item label="资源显示模式">
        <selectResourcesMode v-model="config.resourcesShowMode" />
      </el-form-item>
      <el-form-item label="显示视频时长">
        <el-switch
          v-model="config.showVideoDuration"
          inline-prompt
          active-text="显示"
          inactive-text="关闭"
        />
      </el-form-item>
      <el-form-item v-if="config.resourcesShowMode == 'coverPosterBox'" label="封面海报盒子-信息宽度">
        <el-input-number v-model="config.coverPosterBoxInfoWidth" :min="20" :max="9999" />
      </el-form-item>
      <el-form-item
        v-if="config.resourcesShowMode == 'coverPosterWaterfall'"
        label="封面海报瀑布流-列数"
      >
        <el-input-number v-model="config.coverPosterWaterfallColumn" :min="1" :max="20" />
      </el-form-item>
      <el-form-item label="封面标题对齐方式">
        <el-select v-model="config.coverTitleAlign">
          <el-option label="左对齐" value="left" />
          <el-option label="居中" value="center" />
          <el-option label="右对齐" value="right" />
        </el-select>
      </el-form-item>
      <el-form-item label="资源对齐方式">
        <el-select v-model="config.resourceJustifyContent">
          <el-option label="start" value="flex-start" />
          <el-option label="center " value="center" />
          <el-option label="end" value="flex-end" />
          <el-option label="between" value="space-between" />
          <el-option label="around" value="space-around" />
        </el-select>
      </el-form-item>
      <el-form-item label="详情剧集显示模式">
        <selectDetailsDramaSeriesMode v-model="config.detailsDramaSeriesMode" />
      </el-form-item>
      <el-form-item label="详情显示模式">
        <selectResourceDetailsShowMode v-model="config.resourceDetailsShowMode" />
      </el-form-item>
      <el-form-item label="详情信息显示项">
        <el-checkbox-group v-model="config.detailsVisibleFields">
          <el-checkbox label="副标题" value="subtitle" />
          <el-checkbox label="版号" value="issueNumber" />
          <el-checkbox label="国家" value="country" />
          <el-checkbox label="年份" value="issuingDate" />
          <el-checkbox label="收录时间" value="addTime" />
          <el-checkbox label="清晰度" value="definition" />
          <el-checkbox label="评分" value="score" />
          <el-checkbox label="评星" value="stars" />
        </el-checkbox-group>
      </el-form-item>
      <el-form-item label="封面上显示标签(属性)">
        <el-select v-model="config.coverDisplayTagAttribute" multiple>
          <el-option :label="appLang.attributeTags('definition')" value="definition" />
          <el-option :label="appLang.attributeTags('year')" value="issuingDate" />
          <el-option :label="appLang.attributeTags('country')" value="country" />
          <el-option :label="appLang.attributeTags('starRating')" value="stars" />
          <el-option :label="appLang.attributeTags('score')" value="score" />
          <el-option :label="appLang.attributeTags('hot')" value="hot" />
        </el-select>
      </el-form-item>
    </SettingFieldsScope>
    <slot name="tags" />
    <SettingFieldsScope :disabled="readonly">
      <el-form-item label="标签背景色">
        <div class="color-picker-block">
          <div v-for="(_, index) in config.coverDisplayTagRgbas" :key="index">
            <el-color-picker v-model="config.coverDisplayTagRgbas[index]" show-alpha />
          </div>
          <el-button-group class="color-picker-btn" size="small">
            <el-button icon="Plus" @click="config.coverDisplayTagRgbas.push(getRandomColor())" />
            <el-button icon="Minus" @click="config.coverDisplayTagRgbas.pop()" />
          </el-button-group>
        </div>
      </el-form-item>
      <el-form-item label="标签字体颜色">
        <div class="color-picker-block">
          <div v-for="(_, index) in config.coverDisplayTagColors" :key="index">
            <el-color-picker v-model="config.coverDisplayTagColors[index]" show-alpha />
          </div>
          <el-button-group class="color-picker-btn" size="small">
            <el-button icon="Plus" @click="config.coverDisplayTagColors.push(getRandomColor())" />
            <el-button icon="Minus" @click="config.coverDisplayTagColors.pop()" />
          </el-button-group>
        </div>
      </el-form-item>
      <el-form-item label="标签字体大小">
        <el-input-number v-model="config.coverDisplayTagFontSize" :min="8" :max="24" />
      </el-form-item>
      <el-form-item label="开启显示模块">
        <div class="module-block-group">
          <div class="module-block">
            <el-checkbox
              class="module-block-checkbox"
              v-model="config.casualViewModule"
              label="随便看看"
              border
            />
            <div class="module-block-value-k">
              <label class="module-block-label">显示数量</label>
              <el-input-number v-model="config.casualViewNumber" />
            </div>
          </div>
          <div class="module-block">
            <el-checkbox
              class="module-block-checkbox"
              v-model="config.historyModule"
              label="历史记录"
              border
            />
            <div class="module-block-value-k">
              <label>显示数量</label>
              <el-input-number v-model="config.historyNumber" />
            </div>
          </div>
          <div class="module-block">
            <el-checkbox
              class="module-block-checkbox"
              v-model="config.hotModule"
              label="热门资源"
              border
            />
            <div class="module-block-value-k">
              <label>显示数量</label>
              <el-input-number v-model="config.hotNumber" />
            </div>
          </div>
        </div>
      </el-form-item>
    </SettingFieldsScope>
    <SettingSectionTitle>剧照设置</SettingSectionTitle>
    <SettingFieldsScope :disabled="readonly">
      <el-form-item label="显示剧照">
        <el-switch
          v-model="config.sampleStatus"
          inline-prompt
          active-text="显示"
          inactive-text="关闭"
        />
      </el-form-item>
      <el-form-item label="剧照最大显示数量">
        <el-input-number v-model="config.sampleShowMax" :min="1" :max="100" />
      </el-form-item>
    </SettingFieldsScope>
    <slot name="sample" />
    <SettingSectionTitle>参数设置</SettingSectionTitle>
    <SettingFieldsScope :disabled="readonly">
      <el-form-item label="视频 - 打开方式">
        <el-select v-model="config.openResModeMovies">
          <el-option label="内置" :value="E_resourceOpenMode.Soft" />
          <el-option label="云播" :value="E_resourceOpenMode.CloundPlay" />
          <el-option label="系统" :value="E_resourceOpenMode.System" />
        </el-select>
      </el-form-item>
      <el-form-item label="内置播放器" v-if="config.openResModeMovies === E_resourceOpenMode.Soft">
        <el-select v-model="config.openResModeMovies_SoftType">
          <el-option label="窗口模式" :value="E_resourceOpenMode_SoftType.Windows" />
          <el-option label="弹窗模式" :value="E_resourceOpenMode_SoftType.Dialog" />
        </el-select>
      </el-form-item>
      <el-form-item label="漫画 - 打开方式">
        <el-select v-model="config.openResModeComic">
          <el-option label="内置" :value="E_resourceOpenMode.Soft" />
          <el-option label="系统" :value="E_resourceOpenMode.System" />
        </el-select>
      </el-form-item>
      <el-form-item label="图集 - 打开方式">
        <el-select v-model="config.openResModeAtlas">
          <el-option label="内置" :value="E_resourceOpenMode.Soft" />
          <el-option label="系统" :value="E_resourceOpenMode.System" />
        </el-select>
      </el-form-item>
      <el-form-item label="获取视频预览图关键帧数量">
        <el-input-number v-model="config.videoPreviewImageCount" :min="1" :max="200" />
      </el-form-item>
    </SettingFieldsScope>
    <SettingSectionTitle>演员&导演自定义</SettingSectionTitle>
    <SettingFieldsScope :disabled="readonly">
      <el-form-item label="演员显示文字">
        <el-input v-model="config.performer_Text" />
      </el-form-item>
      <el-form-item label="导演显示文字">
        <el-input v-model="config.director_Text" />
      </el-form-item>
      <el-form-item label="资源数量角标">
        <el-checkbox
          v-model="config.showPerformerResourceCount"
          label="显示演员关联资源数量"
          border
        />
      </el-form-item>
    </SettingFieldsScope>
    <slot name="avatar" />
    <SettingSectionTitle>插件设置</SettingSectionTitle>
    <SettingFieldsScope :disabled="readonly">
      <el-form-item label="Cup插件">
        <el-checkbox v-model="config.plugInUnit_Cup" label="开启演员Cup插件" border />
        <alert-msg color="warning">
          该插件在演员资料中添加Cup选项，并在左边栏出现Cup标签选择。
        </alert-msg>
      </el-form-item>
      <el-form-item label="Cup显示文字">
        <el-input v-model="config.plugInUnit_Cup_Text" />
      </el-form-item>
    </SettingFieldsScope>
    <SettingSectionTitle>封面海报设置</SettingSectionTitle>
    <slot name="poster" />
    <SettingFieldsScope :disabled="readonly">
      <el-form-item label="封面海报显示宽度">
        <el-checkbox v-model="config.coverPosterWidthStatus" label="开启封面海报宽度控制" border />
        <alert-msg color="warning"> 开启该功能，会限定每个资源封面海报的宽度。 </alert-msg>
      </el-form-item>
      <el-form-item label="宽度基数">
        <el-input-number v-model="config.coverPosterWidthBase" />
      </el-form-item>
      <el-form-item label="封面海报显示高度">
        <el-checkbox v-model="config.coverPosterHeightStatus" label="开启封面海报高度控制" border />
        <alert-msg color="warning"> 开启该功能，会限定每个资源封面海报的高度。 </alert-msg>
      </el-form-item>
      <el-form-item label="高度基数">
        <el-input-number v-model="config.coverPosterHeightBase" />
      </el-form-item>
      <el-form-item label="资源间距">
        <el-input-number
          v-model="config.coverPosterGap"
          :precision="1"
          :min="0"
          :max="50"
          :step="0.1"
        />
      </el-form-item>
      <el-form-item label="左右空距">
        <el-input-number v-model="config.contentPadding" :min="0" :max="50" />
      </el-form-item>
    </SettingFieldsScope>
  </div>
</template>
<script setup lang="ts">
import SettingFieldsScope from '@/components/setting/SettingFieldsScope.vue'
import SettingSectionTitle from '@/components/setting/SettingSectionTitle.vue';
import alertMsg from '@/components/com/feedback/alertMsg.vue'
import type { I_config_app } from '@/dataType/config.dataType'
import selectCountry from '@/components/com/form/selectCountry.vue'
import selectDefinition from '@/components/com/form/selectDefinition.vue'
import selectResourceSort from '@/components/com/form/selectResourceSort.vue'
import selectLeftDisplay from '@/components/com/form/selectLeftDisplay.vue'
import selectLeftColumnMode from '@/components/com/form/selectLeftColumnMode.vue'
import selectTagMode from '@/components/com/form/selectTagMode.vue'
import selectResourcesMode from '@/components/com/form/selectResourcesMode.vue'
import selectDetailsDramaSeriesMode from '@/components/com/form/selectDetailsDramaSeriesMode.vue'
import selectResourceDetailsShowMode from '@/components/com/form/selectResourceDetailsShowMode.vue'
import { E_resourceOpenMode, E_resourceOpenMode_SoftType } from '@/dataType/app.dataType'
import { getRandomColor } from '@/assets/tool'
import { AppLang } from '@/language/app.lang'
const appLang = AppLang()
defineProps<{ config: I_config_app; readonly?: boolean }>()
</script>
<style scoped>
.shared-source-hint {
  color: var(--el-text-color-secondary);
  font-size: 13px;
  line-height: 1.6;
  margin: 16px 0;
}
.shared-fields {
  min-width: 0;
}
.shared-fields :deep(.el-select) {
  min-width: 180px;
}
.color-picker-block,
.color-picker-btn,
.module-block,
.module-block-value-k {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}
.module-block-group {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.form-column-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
</style>
