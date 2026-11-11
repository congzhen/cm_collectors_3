<template>
  <section class="about-software settings-subpage">
    <div class="subpage-content">
      <section class="subpage-group" aria-label="使用声明">
        <div class="group-heading"><h3>使用声明</h3><p>免费开源，专注于本地媒体与文件管理。</p></div>
        <div class="subpage-panel">
          <p class="content-text">
          本软件为免费开源软件，主要用于方便有效地管理本地电脑中的视频、图片与文件等。本软件不包含任何内容，所有内容均由用户自行获取并拥有相应版权。请确保您拥有合法权限来管理和使用这些文件。
        </p>
          <p class="content-text">
          本软件按"现状"提供，不提供任何形式的担保，包括但不限于适销性、特定用途适用性和非侵权性。在任何情况下，作者或版权持有者均不对因使用本软件而产生的任何直接、间接、附带、特殊、惩戒性或后果性损害承担责任。
        </p>
          <p class="content-text">
          请勿将本软件用于任何非法活动，使用本软件时产生的所有法律责任均由使用者自行承担。
        </p>
          <p class="content-text">
          欢迎各位使用者根据使用经验对本软件提出宝贵意见和建议。
        </p>
        </div>
      </section>
      <section class="subpage-group" aria-label="项目与反馈">
        <div class="group-heading"><h3>项目与反馈</h3><p>查看项目源码，或通过邮箱提出意见与建议。</p></div>
        <div class="subpage-panel contact-panel">
          <div class="contact-row"><h4>GitHub</h4><p class="contact-info">https://github.com/congzhen/cm_collectors_3</p></div>
          <div class="contact-row"><h4>反馈邮箱</h4><p class="contact-info">Email: czpk673@gmail.com</p></div>
        </div>
      </section>
    </div>
  </section>
</template>
<script lang="ts" setup>
</script>
<style lang="scss" scoped>
@use '../settingsSubpage';
.content-text { color: var(--el-text-color-regular); font-size: 13px; line-height: 1.9; margin-bottom: 16px; &:last-child { margin-bottom: 0; } }
.contact-row {
  display: grid;
  grid-template-columns: 100px minmax(0, 1fr);
  gap: 16px;
  align-items: baseline;
  padding: 16px 0;
  border-bottom: 1px solid var(--el-border-color-extra-light);
  &:first-child { padding-top: 0; }
  &:last-child { padding-bottom: 0; border-bottom: 0; }
  h4 { font-size: 13px; font-weight: 500; color: var(--el-text-color-secondary); }
}
.contact-info { font-size: 13px; line-height: 1.8; color: var(--el-text-color-primary); overflow-wrap: anywhere; user-select: text; }
@media (max-width: 700px) { .contact-row { grid-template-columns: minmax(0, 1fr); gap: 6px; } }
</style>
