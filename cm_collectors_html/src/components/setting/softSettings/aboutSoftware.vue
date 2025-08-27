<template>
  <div class="about-software">
    <div class="content-wrapper">
      <h3 class="section-title">声明:</h3>
      <div class="content-section">
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

      <h3 class="section-title">联系方式:</h3>
      <div class="content-section">
        <p class="contact-info">Email: czpk673@gmail.com</p>
        <!--
        <div class="telegram-section">
          <div class="qr-code-container">
            <img src="/tg.png" alt="Telegram" width="200" height="200" />
          </div>
          <div class="telegram-info">
            <div class="telegram-label">TG讨论群</div>
            <div class="telegram-link">https://t.me/+rFdWuJ8si_IzNTc1</div>
          </div>
        </div>
      -->
      </div>
    </div>
  </div>
</template>
<script lang="ts" setup>
</script>
<style lang="scss" scoped>
.about-software {
  padding: 20px;
  min-height: 100%;
  background-color: rgba(0, 0, 0, 0.1);
  border-radius: 8px;

  .content-wrapper {
    max-width: 800px;
    margin: 0 auto;
  }

  .section-title {
    font-size: 18px;
    font-weight: bold;
    color: #409eff;
    margin: 20px 0 10px 0;
    padding-bottom: 5px;
    border-bottom: 1px solid rgba(64, 158, 255, 0.3);
  }

  .content-section {
    padding: 10px 0;
  }

  .content-text {
    font-size: 14px;
    line-height: 1.8;
    margin-bottom: 15px;
    text-align: justify;
    color: #f2f2f2;
    text-indent: 2em;
  }

  .contact-info {
    font-size: 14px;
    line-height: 1.6;
    color: #f2f2f2;
    padding: 10px 15px;
    background-color: rgba(64, 158, 255, 0.1);
    border-radius: 4px;
    display: inline-block;
  }

  .telegram-section {
    display: flex;
    align-items: center;
    gap: 20px;
    padding: 20px;
    background-color: rgba(64, 158, 255, 0.1);
    border-radius: 8px;
    margin-top: 15px;
  }

  .qr-code-container {
    flex-shrink: 0;
    border: 2px solid #409eff;
    border-radius: 8px;
    padding: 5px;
    background-color: white;
  }

  .telegram-info {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .telegram-label {
    font-size: 16px;
    font-weight: bold;
    color: #409eff;
  }

  .telegram-link {
    font-size: 14px;
    color: #f2f2f2;
    word-break: break-all;
    background-color: rgba(0, 0, 0, 0.2);
    padding: 8px 12px;
    border-radius: 4px;
    cursor: pointer;
    transition: background-color 0.3s;

    &:hover {
      background-color: rgba(64, 158, 255, 0.3);
    }
  }
}
</style>
