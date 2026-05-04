<template>
  <aside class="library-rail">
    <div class="rail-title">
      <span>LIBRARY</span>
      <strong>{{ filesBasesList.length }}</strong>
    </div>
    <div class="rail-list">
      <button v-for="filesBases in filesBasesList" :key="filesBases.id" class="rail-item"
        :class="{ active: filesBases.id == currentFilesBasesId }" @click="emits('select', filesBases)">
        <span class="rail-mark">{{ filesBases.name.slice(0, 1).toUpperCase() }}</span>
        <span class="rail-name">{{ filesBases.name }}</span>
      </button>
    </div>
    <div class="rail-stats">
      <div>
        <span>标签组</span>
        <strong>{{ tagClassCount }}</strong>
      </div>
      <div>
        <span>演员库</span>
        <strong>{{ performerBasesCount }}</strong>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import type { PropType } from 'vue'
import type { I_filesBases } from '@/dataType/filesBases.dataType'

defineProps({
  filesBasesList: {
    type: Array as PropType<I_filesBases[]>,
    default: () => [],
  },
  currentFilesBasesId: {
    type: String,
    default: '',
  },
  tagClassCount: {
    type: Number,
    default: 0,
  },
  performerBasesCount: {
    type: Number,
    default: 0,
  },
})

const emits = defineEmits<{
  select: [filesBases: I_filesBases]
}>()
</script>

<style lang="scss" scoped>
.library-rail {
  min-height: 0;
  padding: 10px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  border: 1px solid var(--home-border);
  border-radius: 8px;
  background: var(--home-panel-bg);
  box-sizing: border-box;
}

.rail-title {
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: var(--home-text-muted);
  letter-spacing: 0;

  span {
    font-size: 12px;
    font-weight: 800;
  }

  strong {
    width: 32px;
    height: 32px;
    display: grid;
    place-items: center;
    border-radius: 8px;
    color: var(--home-accent);
    background: var(--home-accent-soft);
  }
}

.rail-list {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
  overflow: auto;
}

.rail-item {
  width: 100%;
  height: 42px;
  padding: 6px 8px;
  display: flex;
  align-items: center;
  gap: 10px;
  border: 1px solid var(--home-border);
  border-radius: 8px;
  color: var(--home-text);
  background: var(--home-panel-soft-bg);
  cursor: pointer;
  text-align: left;
  transition: all 0.18s ease;

  &:hover {
    border-color: var(--home-accent);
    color: var(--home-accent);
  }

  &.active {
    color: #ffffff;
    border-color: var(--home-accent);
    background: var(--home-accent);
  }
}

.rail-mark {
  width: 30px;
  height: 30px;
  flex-shrink: 0;
  display: grid;
  place-items: center;
  border-radius: 6px;
  color: #ffffff;
  background: var(--home-accent);
  font-weight: 800;
}

.rail-name {
  min-width: 0;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  font-weight: 700;
}

.rail-stats {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;

  div {
    padding: 10px;
    border-radius: 8px;
    background: var(--home-panel-soft-bg);
    border: 1px solid var(--home-border);
  }

  span {
    display: block;
    color: var(--home-text-muted);
    line-height: 16px;
  }

  strong {
    color: var(--home-text);
    font-size: 22px;
    line-height: 28px;
  }
}

@media (max-width: 1100px) {
  .library-rail {
    min-height: 132px;
  }

  .rail-list {
    flex-direction: row;
    overflow-x: auto;
  }

  .rail-item {
    min-width: 150px;
  }
}
</style>
