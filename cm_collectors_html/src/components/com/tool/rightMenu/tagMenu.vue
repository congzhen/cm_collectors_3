<template>
  <div class="context-menu-host" @contextmenu="handleContextMenu">
    <slot></slot>

    <teleport to="body">
      <div v-if="isVisible" ref="contextMenuRef" class="context-menu" :style="menuStyle" @contextmenu.prevent>
        <!-- 标题行 -->
        <div v-if="props.title" class="context-menu__title">
          {{ props.title }}
        </div>

        <div v-for="(item, index) in props.items" :key="index" :class="getItemClass(item)"
          @click="handleItemClick(item)" @mouseenter="handleItemHover(item, index)">
          <template v-if="!item.separator">
            <span v-if="item.icon" class="context-menu-item__icon">
              <el-icon v-if="isElementPlusIcon(item.icon)">
                <component :is="item.icon" />
              </el-icon>
              <i v-else :class="item.icon"></i>
            </span>
            <span class="context-menu-item__label">{{ item.label }}</span>
            <span v-if="item.children && item.children.length" class="context-menu-item__arrow">
              ▶
            </span>
            <span v-else-if="item.shortcut" class="context-menu-item__shortcut">
              {{ item.shortcut }}
            </span>
          </template>
        </div>

        <!-- 子菜单 -->
        <teleport to="body">
          <div v-if="activeSubmenu" class="context-submenu" :style="submenuStyle">
            <div v-for="(subItem, subIndex) in activeSubmenu.children" :key="subIndex" :class="getItemClass(subItem)"
              @click="handleItemClick(subItem)">
              <template v-if="!subItem.separator">
                <span v-if="subItem.icon" class="context-menu-item__icon">
                  <el-icon v-if="isElementPlusIcon(subItem.icon)">
                    <component :is="subItem.icon" />
                  </el-icon>
                  <i v-else :class="subItem.icon"></i>
                </span>
                <span class="context-menu-item__label">{{ subItem.label }}</span>
                <span v-if="subItem.shortcut" class="context-menu-item__shortcut">
                  {{ subItem.shortcut }}
                </span>
              </template>
            </div>
          </div>
        </teleport>
      </div>
    </teleport>
  </div>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted, onUnmounted, type Ref, type PropType, nextTick } from 'vue'
import menuManager from './menuManager'


// 定义菜单项接口
interface ContextMenuItem {
  label?: string // 菜单项标签
  icon?: string // 菜单项图标
  shortcut?: string // 菜单项快捷键
  disabled?: boolean // 菜单项是否禁用
  separator?: boolean  // 分割线
  children?: ContextMenuItem[] // 子菜单项
  handler?: () => void // 菜单项点击处理函数
  onClick?: () => void // 菜单项点击处理函数
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  [key: string]: any // 其他属性
}

const props = defineProps({
  width: {
    type: String,
    default: '200px',
  },
  items: {
    type: Array as PropType<ContextMenuItem[]>,
    default: () => [],
  },
  title: {
    type: String,
    default: '',
  },
})

const emit = defineEmits<{
  (e: 'contextmenu', event: MouseEvent): void
}>()

// 状态管理
const isVisible = ref(false)
const position = ref({ x: 0, y: 0 })
const activeSubmenu: Ref<ContextMenuItem | null> = ref(null)
const activeSubmenuIndex = ref<number | null>(null)
const contextMenuRef: Ref<HTMLElement | null> = ref(null)
// 判断是否为 Element Plus 图标
const isElementPlusIcon = (icon: string | undefined) => {
  // Element Plus 图标通常是大写字母开头的单词，如 'Edit', 'Delete', 'Plus'
  if (!icon) return false
  return /^[A-Z]/.test(icon)
}

// 计算菜单位置样式
const menuStyle = computed(() => {
  return {
    top: `${position.value.y}px`,
    left: `${position.value.x}px`,
    width: props.width
  }
})

// 计算子菜单位置样式
const submenuStyle = computed(() => {
  if (!activeSubmenu.value) return {}

  // 默认位置
  let top = position.value.y
  let left = position.value.x + parseInt(props.width)

  // 检查子菜单是否超出右侧边界
  if (left + parseInt(props.width) > window.innerWidth) {
    // 如果超出右侧边界，则显示在主菜单左侧
    left = position.value.x - parseInt(props.width)
  }

  // 检查子菜单是否超出底部边界
  // 这里使用估计高度，实际应该根据子菜单项数计算
  const estimatedHeight = activeSubmenu.value.children ? activeSubmenu.value.children.length * 36 + 30 : 200
  if (top + estimatedHeight > window.innerHeight) {
    // 向上调整位置
    top = Math.max(0, window.innerHeight - estimatedHeight)
  }

  return {
    top: `${top}px`,
    left: `${left}px`,
    minWidth: props.width,
    maxWidth: props.width
  }
})

// 处理右键事件
const handleContextMenu = (event: MouseEvent) => {
  event.preventDefault()
  emit('contextmenu', event)

  // 注册当前菜单到全局管理器
  menuManager.registerMenu(hideMenu)

  // 初始位置
  const x = Math.min(event.clientX, window.innerWidth - parseInt(props.width))

  // 先设置菜单位置
  position.value = { x, y: event.clientY }
  isVisible.value = true
  activeSubmenu.value = null
  activeSubmenuIndex.value = null

  // 使用 nextTick 确保 DOM 已更新后再计算位置
  nextTick(() => {
    if (contextMenuRef.value) {
      const menuRect = contextMenuRef.value.getBoundingClientRect()
      const menuHeight = menuRect.height

      // 计算最终y坐标，确保菜单不会超出视口
      let finalY = event.clientY

      // 如果菜单向下展开会超出底部边界
      if (event.clientY + menuHeight > window.innerHeight) {
        // 尝试向上展开
        const upwardY = event.clientY - menuHeight

        // 如果向上展开会超出顶部边界
        if (upwardY < 0) {
          // 最终选择保留更多菜单项的方案
          // 如果向下展开能显示更多菜单项
          if (window.innerHeight - event.clientY > event.clientY) {
            // 向下展开，但限制最大高度
            finalY = event.clientY
            contextMenuRef.value.style.maxHeight = (window.innerHeight - event.clientY - 10) + 'px'
            contextMenuRef.value.style.overflowY = 'auto'
          } else {
            // 向上展开，但限制最大高度
            finalY = 10
            contextMenuRef.value.style.maxHeight = (event.clientY - 20) + 'px'
            contextMenuRef.value.style.overflowY = 'auto'
          }
        } else {
          // 正常向上展开
          finalY = upwardY
        }
      }

      position.value.y = finalY
    }
  })
}

// 处理菜单项点击
const handleItemClick = (item: ContextMenuItem) => {
  if (item.disabled || item.separator) return

  hideMenu()

  if (item.handler) {
    item.handler()
  } else if (item.onClick) {
    item.onClick()
  }
}

// 处理菜单项悬停（显示子菜单）
const handleItemHover = (item: ContextMenuItem, index: number) => {
  if (item.disabled || item.separator) {
    activeSubmenu.value = null
    activeSubmenuIndex.value = null
    return
  }

  if (item.children && item.children.length) {
    activeSubmenu.value = item
    activeSubmenuIndex.value = index
  } else {
    activeSubmenu.value = null
    activeSubmenuIndex.value = null
  }
}

// 获取菜单项的类名
const getItemClass = (item: ContextMenuItem) => {
  return {
    'context-menu-item': true,
    'context-menu-item--disabled': item.disabled,
    'context-menu-item--separator': item.separator,
    'context-menu-item--has-children': item.children && item.children.length,
    'context-menu-item--active': activeSubmenu.value &&
      item.children &&
      item.children.length &&
      activeSubmenu.value === item
  }
}

// 隐藏菜单
const hideMenu = () => {
  isVisible.value = false
  activeSubmenu.value = null
  activeSubmenuIndex.value = null
  menuManager.unregisterMenu()

  // 清除可能设置的滚动样式
  if (contextMenuRef.value) {
    contextMenuRef.value.style.maxHeight = ''
    contextMenuRef.value.style.overflowY = ''
  }
}

// 处理点击外部区域
const handleClickOutside = (event: MouseEvent) => {
  if (!isVisible.value) return
  const menuEl = contextMenuRef.value
  if (menuEl && !menuEl.contains(event.target as Node)) {
    hideMenu()
  }
}

// 组件挂载时添加事件监听器
onMounted(() => {
  document.addEventListener('click', handleClickOutside as EventListener, true)
  document.addEventListener('scroll', hideMenu)
})

// 组件卸载时移除事件监听器
onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside as EventListener, true)
  document.removeEventListener('scroll', hideMenu)
  // 确保在组件卸载时从管理器中移除
  menuManager.unregisterMenu()
})

// 暴露方法给父组件
defineExpose({
  show: handleContextMenu,
  hide: hideMenu
})
</script>

<style scoped>
.context-menu-host {
  display: contents;
}

.context-menu {
  position: fixed;
  z-index: 9999;
  background: #ffffff;
  border-radius: 6px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  padding: 4px 0;
  min-width: 200px;
  font-size: 14px;
  color: #333;
  border: 1px solid #e0e0e0;
  user-select: none;
}

.context-menu__title {
  padding: 8px 12px;
  font-weight: bold;
  border-bottom: 1px solid #e0e0e0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.context-menu-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  cursor: pointer;
  position: relative;
}

.context-menu-item:hover:not(.context-menu-item--disabled):not(.context-menu-item--separator) {
  background-color: #f0f0f0;
}

.context-menu-item--disabled {
  color: #999;
  cursor: not-allowed;
}

.context-menu-item--separator {
  height: 1px;
  background-color: #e0e0e0;
  margin: 4px 0;
  padding: 0;
  cursor: default;
}

.context-menu-item__icon {
  width: 16px;
  margin: 0 8px 2px 0;
  text-align: center;
  display: flex;
  align-items: center;
}

.context-menu-item__label {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.context-menu-item__shortcut {
  color: #999;
  font-size: 12px;
}

.context-menu-item__arrow {
  color: #999;
  font-size: 12px;
}

.context-submenu {
  position: fixed;
  background: #ffffff;
  border-radius: 6px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  padding: 4px 0;
  min-width: 200px;
  border: 1px solid #e0e0e0;
  z-index: 10000;
}
</style>
