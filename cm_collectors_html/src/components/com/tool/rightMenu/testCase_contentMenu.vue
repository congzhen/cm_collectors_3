<template>
  <div class="context-menu-demo">
    <h2>自定义右键菜单示例</h2>

    <div class="demo-grid">
      <!-- 基本用法 -->
      <div class="demo-section">
        <h3>基本用法</h3>
        <CustomContextMenu :items="basicMenuItems" @contextmenu="onContextMenu">
          <div class="demo-area">
            <p>在此区域右键点击显示菜单</p>
          </div>
        </CustomContextMenu>
      </div>

      <!-- 带子菜单 -->
      <div class="demo-section">
        <h3>嵌套菜单</h3>
        <CustomContextMenu :items="nestedMenuItems" @contextmenu="onContextMenu">
          <div class="demo-area nested">
            <p>在此区域右键点击显示嵌套菜单</p>
          </div>
        </CustomContextMenu>
      </div>

      <!-- 列表项 -->
      <div class="demo-section full-width">
        <h3>列表项右键菜单</h3>
        <ul class="item-list">
          <li v-for="item in listItems" :key="item.id" @contextmenu="showItemMenu($event, item)" class="list-item">
            {{ item.name }}
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import CustomContextMenu from './contentMenu.vue'

// 基本菜单项
const basicMenuItems = [
  {
    label: '新建',
    icon: 'icon-plus',
    shortcut: 'Ctrl+N',
    handler: () => {
      alert('执行新建操作')
    }
  },
  {
    label: '打开',
    icon: 'icon-folder',
    handler: () => {
      alert('执行打开操作')
    }
  },
  {
    separator: true
  },
  {
    label: '编辑',
    icon: 'icon-edit',
    handler: () => {
      alert('执行编辑操作')
    }
  },
  {
    label: '删除',
    icon: 'icon-delete',
    disabled: true,
    handler: () => {
      alert('执行删除操作')
    }
  },
  {
    separator: true
  },
  {
    label: '属性',
    icon: 'icon-info',
    handler: () => {
      alert('查看属性')
    }
  }
]

// 嵌套菜单项
const nestedMenuItems = [
  {
    label: '文件',
    icon: 'icon-file',
    children: [
      {
        label: '新建',
        handler: () => alert('新建文件')
      },
      {
        label: '新建窗口',
        handler: () => alert('新建窗口')
      }
    ]
  },
  {
    label: '编辑',
    icon: 'icon-edit',
    children: [
      {
        label: '撤销',
        shortcut: 'Ctrl+Z',
        handler: () => alert('撤销操作')
      },
      {
        label: '重做',
        shortcut: 'Ctrl+Y',
        handler: () => alert('重做操作')
      },
      {
        separator: true
      },
      {
        label: '复制',
        shortcut: 'Ctrl+C',
        handler: () => alert('复制')
      },
      {
        label: '粘贴',
        shortcut: 'Ctrl+V',
        handler: () => alert('粘贴')
      }
    ]
  },
  {
    separator: true
  },
  {
    label: '视图',
    icon: 'icon-view',
    children: [
      {
        label: '放大',
        handler: () => alert('放大')
      },
      {
        label: '缩小',
        handler: () => alert('缩小')
      },
      {
        label: '重置',
        handler: () => alert('重置')
      }
    ]
  }
]

// 列表项数据
const listItems = ref([
  { id: 1, name: '文档1.txt', type: 'file' },
  { id: 2, name: '图片目录', type: 'folder' },
  { id: 3, name: '视频.mp4', type: 'video' },
  { id: 4, name: '音频.mp3', type: 'audio' }
])

// 事件处理
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const onContextMenu = (event: any) => {
  console.log('显示右键菜单', event)
}

// 显示列表项菜单
const showItemMenu = (event: { preventDefault: () => void }, item: { name: string | undefined; id: number }) => {
  event.preventDefault()

  const itemMenuItems = [
    {
      label: `打开 ${item.name}`,
      icon: 'icon-open',
      handler: () => {
        alert(`打开 ${item.name}`)
      }
    },
    {
      label: '重命名',
      icon: 'icon-edit',
      handler: () => {
        const newName = prompt('请输入新名称:', item.name)
        if (newName) {
          item.name = newName
        }
      }
    },
    {
      separator: true
    },
    {
      label: '删除',
      icon: 'icon-delete',
      handler: () => {
        if (confirm(`确定要删除 ${item.name} 吗？`)) {
          listItems.value = listItems.value.filter(i => i.id !== item.id)
        }
      }
    }
  ]

  // 创建临时菜单组件
  const tempMenu = document.createElement('div')
  document.body.appendChild(tempMenu)

  import('./contentMenu.vue').then(module => {
    const { createApp } = require('vue')
    const CustomContextMenu = module.default

    const app = createApp(CustomContextMenu, {
      items: itemMenuItems
    })

    app.mount(tempMenu)

    // 触发菜单显示
    const menuComponent = app._instance.proxy
    menuComponent.show(event)

    // 菜单关闭后清理
    const cleanup = () => {
      app.unmount()
      document.body.removeChild(tempMenu)
      document.removeEventListener('click', cleanup)
    }

    setTimeout(() => {
      document.addEventListener('click', cleanup)
    }, 0)
  })
}
</script>

<style scoped>
.context-menu-demo {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.demo-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 20px;
  margin-top: 20px;
}

.demo-section.full-width {
  grid-column: 1 / -1;
}

.demo-section h3 {
  margin-top: 0;
}

.demo-area {
  height: 150px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #f8f9fa;
  border: 2px dashed #6c757d;
  border-radius: 4px;
  cursor: context-menu;
  user-select: none;
}

.demo-area:hover {
  background-color: #e9ecef;
}

.demo-area.nested {
  background-color: #fff3cd;
  border-color: #ffc107;
}

.demo-area.nested:hover {
  background-color: #ffeaa7;
}

.item-list {
  list-style: none;
  padding: 0;
  margin: 0;
  border: 1px solid #dee2e6;
  border-radius: 4px;
}

.list-item {
  padding: 12px 16px;
  border-bottom: 1px solid #dee2e6;
  cursor: context-menu;
  user-select: none;
}

.list-item:hover {
  background-color: #f8f9fa;
}

.list-item:last-child {
  border-bottom: none;
}
</style>
