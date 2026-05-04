<template>
  <section class="stage-hero">
    <div class="stage-title">
      <span class="eyebrow">RESOURCES VIEW</span>
      <div class="title-line">
        <h1>{{ title }}</h1>
        <p>{{ subtitle }}</p>
      </div>
    </div>
    <div class="stage-actions">
      <el-button size="small" :type="filterVisible ? 'primary' : 'default'" icon="Filter" plain
        @click="emits('toggleFilter')" />
    </div>
  </section>
</template>

<script setup lang="ts">

defineProps({
  title: {
    type: String,
    default: '',
  },
  subtitle: {
    type: String,
    default: '',
  },
  filterVisible: {
    type: Boolean,
    default: true,
  },
})

const emits = defineEmits<{
  toggleFilter: []
}>()
</script>

<style lang="scss" scoped>
.stage-hero {
  min-height: 48px;
  padding: 8px 12px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border: 1px solid var(--home-border);
  border-radius: 8px;
  color: var(--home-text);
  background:
    linear-gradient(90deg, var(--home-panel-bg), var(--home-panel-soft-bg));
  box-sizing: border-box;
}

.stage-title {
  min-width: 0;

  .eyebrow {
    display: block;
    color: var(--home-accent);
    font-size: 11px;
    font-weight: 800;
    line-height: 14px;
  }

  .title-line {
    display: flex;
    align-items: baseline;
    gap: 10px;
    min-width: 0;
  }

  h1 {
    flex-shrink: 1;
    max-width: 58vw;
    margin: 0;
    color: var(--home-text);
    font-size: 18px;
    line-height: 22px;
    font-weight: 800;
    letter-spacing: 0;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  p {
    flex-shrink: 0;
    margin: 0;
    color: var(--home-text-muted);
    line-height: 15px;
  }
}

.stage-actions {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 8px;

}

@media (max-width: 1100px) {
  .stage-hero {
    align-items: flex-start;
    flex-direction: column;
  }

  .stage-title h1 {
    max-width: 100%;
    font-size: 24px;
    line-height: 30px;
  }

  .title-line {
    align-items: flex-start;
    flex-direction: column;
    gap: 2px;
  }

  .stage-actions {
    justify-content: flex-start;
    flex-wrap: wrap;
  }
}
</style>
