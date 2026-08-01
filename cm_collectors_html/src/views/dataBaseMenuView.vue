<template>
  <div v-if="appearanceStyle === 'modern'" class="database-menu-modern"
    :class="{ 'database-menu-modern--bright': isBrightTheme }">
    <button v-show="canScrollLeft" class="database-menu-scroll database-menu-scroll--left" type="button"
      aria-label="向左滚动数据库列表" @click="scrollMenu(-1)">
      <el-icon><ArrowLeft /></el-icon>
    </button>
    <div ref="modernMenuRef" class="database-menu-modern-track" role="tablist" aria-label="数据库切换"
      @scroll="updateScrollStatus" @wheel="horizontalWheelHandle">
      <button v-for="filesBases in store.filesBasesStoreData.filesBasesStatus" :key="filesBases.id"
        class="database-menu-modern-item" :class="{ active: filesBases.id === activeDataBase }" type="button"
        role="tab" :aria-selected="filesBases.id === activeDataBase" @click="changeDataBaseHandle(filesBases, $event)">
        {{ filesBases.name }}
      </button>
    </div>
    <button v-show="canScrollRight" class="database-menu-scroll database-menu-scroll--right" type="button"
      aria-label="向右滚动数据库列表" @click="scrollMenu(1)">
      <el-icon><ArrowRight /></el-icon>
    </button>
  </div>
  <div v-else class="database-menu">
    <el-menu :default-active="activeDataBase" mode="horizontal" text-color="#fff" active-text-color="#ffd04b">
      <el-menu-item :index="filesBases.id" v-for="filesBases, key in store.filesBasesStoreData.filesBasesStatus"
        :key="key" @click="changeDataBaseHandle(filesBases)">
        {{ filesBases.name }}
      </el-menu-item>
    </el-menu>
  </div>
</template>
<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { filesBasesStoreData } from '@/storeData/filesBases.storeData';
import { appStoreData } from '@/storeData/app.storeData';
import type { I_filesBases } from '@/dataType/filesBases.dataType';
const store = {
  filesBasesStoreData: filesBasesStoreData(),
  appStoreData: appStoreData(),
}
const emits = defineEmits(['selectFilesBase']);

const activeDataBase = ref(store.filesBasesStoreData.filesBasesFirst?.id || '')
const modernMenuRef = ref<HTMLElement>()
const canScrollLeft = ref(false)
const canScrollRight = ref(false)
let menuResizeObserver: ResizeObserver | undefined

const appearanceStyle = computed(() =>
  store.appStoreData.appConfig.appearanceStyle
  || store.appStoreData.appConfig.headerStyle
  || 'modern'
)
const isBrightTheme = computed(() => store.appStoreData.appConfig.theme === 'bright')

const updateScrollStatus = () => {
  const menu = modernMenuRef.value
  if (!menu) return
  canScrollLeft.value = menu.scrollLeft > 2
  canScrollRight.value = menu.scrollLeft + menu.clientWidth < menu.scrollWidth - 2
}

const scrollMenu = (direction: -1 | 1) => {
  modernMenuRef.value?.scrollBy({ left: direction * 260, behavior: 'smooth' })
}

const horizontalWheelHandle = (event: WheelEvent) => {
  const menu = modernMenuRef.value
  if (!menu || menu.scrollWidth <= menu.clientWidth) return
  if (Math.abs(event.deltaY) <= Math.abs(event.deltaX)) return
  event.preventDefault()
  menu.scrollLeft += event.deltaY
}

const changeDataBaseHandle = (filesBases: I_filesBases, event?: MouseEvent) => {
  if (filesBases.id != activeDataBase.value) {
    activeDataBase.value = filesBases.id;
    emits('selectFilesBase', filesBases);
  }
  ;(event?.currentTarget as HTMLElement | null)?.scrollIntoView({ behavior: 'smooth', block: 'nearest', inline: 'center' })
}

watch(() => store.filesBasesStoreData.filesBasesStatus.length, async () => {
  await nextTick()
  updateScrollStatus()
})

onMounted(async () => {
  await nextTick()
  if (modernMenuRef.value) {
    menuResizeObserver = new ResizeObserver(updateScrollStatus)
    menuResizeObserver.observe(modernMenuRef.value)
  }
  updateScrollStatus()
})

onBeforeUnmount(() => menuResizeObserver?.disconnect())

</script>
<style lang="scss" scoped>
.database-menu {
  height: 31px;
  padding: 5px 0;

  :deep(.el-menu--horizontal) {
    height: auto;
  }

  :deep(.el-menu--horizontal>.el-menu-item) {
    height: 14px;
    line-height: 16px;
    padding-top: 12px;
    padding-bottom: 16px;

    &:hover {
      background-color: unset;
      color: #409EFF;
      user-select: none;
    }
  }

}
</style>

<style lang="scss" scoped>
.database-menu-modern {
  --database-menu-bg: #1f1f1f;
  --database-menu-border: rgba(255, 255, 255, 0.1);
  --database-menu-text: #a8abb2;
  --database-menu-text-active: #37c6ca;
  --database-menu-hover: rgba(255, 255, 255, 0.045);
  --database-menu-active: rgba(55, 198, 202, 0.1);
  --database-menu-shadow: 0 5px 12px rgba(0, 0, 0, 0.18);

  position: relative;
  z-index: 2;
  height: 44px;
  padding: 0 7px;
  display: flex;
  align-items: stretch;
  box-sizing: border-box;
  overflow: hidden;
  background: var(--database-menu-bg);
  border: 0;
  box-shadow: var(--database-menu-shadow);

  &--bright {
    --database-menu-bg: #ffffff;
    --database-menu-border: #e2e6eb;
    --database-menu-text: #606b75;
    --database-menu-text-active: #159fa1;
    --database-menu-hover: #f5f7f9;
    --database-menu-active: rgba(21, 159, 161, 0.09);
    --database-menu-shadow: 0 4px 12px rgba(45, 58, 70, 0.09);
  }
}

.database-menu-modern-track {
  min-width: 0;
  flex: 1;
  display: flex;
  align-items: stretch;
  gap: 4px;
  overflow-x: auto;
  overflow-y: hidden;
  scrollbar-width: none;
  scroll-behavior: smooth;

  &::-webkit-scrollbar {
    display: none;
  }
}

.database-menu-modern-item {
  position: relative;
  min-width: max-content;
  height: 100%;
  padding: 0 17px;
  flex: 0 0 auto;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 6px 6px 0 0;
  color: var(--database-menu-text);
  background: transparent;
  font: inherit;
  font-size: 13px;
  font-weight: 500;
  white-space: nowrap;
  cursor: pointer;
  transition: color 0.16s ease, background-color 0.16s ease;

  &::after {
    content: '';
    position: absolute;
    right: 12px;
    bottom: 0;
    left: 12px;
    height: 2px;
    border-radius: 2px 2px 0 0;
    background: transparent;
    transform: scaleX(0.4);
    transition: background-color 0.16s ease, transform 0.16s ease;
  }

  &:hover {
    color: var(--database-menu-text-active);
    background: var(--database-menu-hover);
  }

  &.active {
    color: var(--database-menu-text-active);
    background: var(--database-menu-active);

    &::after {
      background: var(--database-menu-text-active);
      transform: scaleX(1);
    }
  }
}

.database-menu-scroll {
  position: absolute;
  z-index: 2;
  top: 0;
  bottom: 0;
  width: 34px;
  padding: 0;
  display: grid;
  place-items: center;
  border: 0;
  color: var(--database-menu-text);
  background: var(--database-menu-bg);
  cursor: pointer;

  &::after {
    content: '';
    position: absolute;
    top: 0;
    bottom: 0;
    width: 18px;
    pointer-events: none;
  }

  &:hover {
    color: var(--database-menu-text-active);
  }

  &--left {
    left: 0;

    &::after {
      left: 100%;
      background: linear-gradient(90deg, var(--database-menu-bg), transparent);
    }
  }

  &--right {
    right: 0;

    &::after {
      right: 100%;
      background: linear-gradient(270deg, var(--database-menu-bg), transparent);
    }
  }
}
</style>
