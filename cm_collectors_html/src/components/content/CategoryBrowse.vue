<template>
  <section class="category-browse" v-loading="loading">
    <header class="category-toolbar">
      <div>
        <div class="category-breadcrumb" v-if="selected">
          <button type="button" @click="back">分类列表</button><span>/</span>
          <strong>{{ selected.name }}</strong
          ><span v-if="!all">/ 第 {{ selected.part }} 组</span>
        </div>
        <strong v-else>{{ sourceLabel }}</strong>
        <small>跟随当前筛选{{ randomSort ? ' · 随机排序时按收录时间稳定分组' : '' }}</small>
      </div>
      <div>
        <el-radio-group
          v-if="selected && settings.categoryBrowseGroupLimit > 0 && (selected.parts > 1 || all)"
          :model-value="all"
          size="small"
          @update:model-value="enter(selected, $event === true)"
        >
          <el-radio-button :value="false">本组</el-radio-button
          ><el-radio-button :value="true">全部</el-radio-button>
        </el-radio-group>
        <el-button text size="small" @click="emit('exit')">返回资源浏览</el-button>
      </div>
    </header>
    <div v-if="error" class="category-error">
      <el-alert :title="error" type="error" :closable="false" /><el-button @click="load"
        >重试</el-button
      >
    </div>
    <template v-else>
      <div v-if="selected" class="category-resources">
        <contentList
          v-if="loading || members.length"
          :data-list="members"
          plain
          @select-resources="emit('select', $event)"
        />
        <el-empty
          v-if="!loading && !members.length"
          description="当前筛选下此组暂无资源，可返回分类列表重新查看"
        />
      </div>
      <el-scrollbar v-else class="category-scroll">
        <div class="category-grid">
          <CategoryBrowseCard
            v-for="group in groups"
            :key="`${group.categoryId}:${group.part}`"
            :group="group"
            :query="query"
            :cover-style="settings.categoryBrowseCoverStyle"
            :expand-style="settings.categoryBrowseExpandStyle"
            @enter="enter"
          />
        </div>
        <el-empty v-if="!loading && !groups.length" description="当前筛选下没有资源" />
      </el-scrollbar>
    </template>
    <footer v-if="!externalFooter">
      <span>{{
        selected
          ? `共 ${memberTotal} 个资源${all ? '' : ` · 本组 ${members.length} 个`}`
          : `共 ${categoryTotal} 个分类 · ${total} 个展示组 · ${resourceTotal} 个资源`
      }}</span>
      <el-pagination
        v-if="!selected || all"
        v-model:current-page="page"
        :total="selected ? memberTotal : total"
        :page-size="pageSize"
        layout="prev,pager,next,jumper"
        :pager-count="5"
        small
      />
    </footer>
  </section>
</template>
<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { appStoreData } from '@/storeData/app.storeData'
import { searchStoreData } from '@/storeData/search.storeData'
import {
  categoryBrowseServer,
  categorySettings,
  type CategoryGroup,
  type CategoryQuery,
} from '@/server/categoryBrowse.server'
import type { I_resource } from '@/dataType/resource.dataType'
import { E_searchSort } from '@/dataType/search.dataType'
import CategoryBrowseCard from './CategoryBrowseCard.vue'
import contentList from './contentList.vue'
defineProps<{ externalFooter?: boolean }>()
const emit = defineEmits<{ select: [resource: I_resource]; clearSelection: []; exit: [] }>()
const store = appStoreData(),
  search = searchStoreData()
const settings = computed(() => categorySettings(store.currentConfigApp))
const sourceLabel = computed(() =>
  settings.value.categoryBrowseSource === 'performer'
    ? '按演员浏览'
    : `按${store.currentTagClass?.find((item) => `tag:${item.id}` === settings.value.categoryBrowseSource)?.name || '标签分类'}浏览`,
)
const randomSort = computed(() => search.searchData.sort === E_searchSort.Random)
const page = ref(1),
  pageSize = computed(() => Math.min(100, Math.max(1, store.currentConfigApp.pageLimit || 24)))
const groups = ref<CategoryGroup[]>([]),
  members = ref<I_resource[]>([]),
  selected = ref<CategoryGroup>(),
  all = ref(false)
const total = ref(0),
  resourceTotal = ref(0),
  categoryTotal = ref(0),
  memberTotal = ref(0),
  loading = ref(false),
  error = ref('')
let version = 0
let categoryPage = 1
let durationTimer: ReturnType<typeof setTimeout> | undefined
let lastMembersQuery: CategoryQuery | undefined
const query = computed<CategoryQuery>(() => ({
  filesBasesId: store.currentFilesBases.id,
  source: settings.value.categoryBrowseSource,
  categorySort: settings.value.categoryBrowseSort,
  groupLimit: settings.value.categoryBrowseGroupLimit,
  searchData: search.searchData,
  page: page.value,
  limit: pageSize.value,
}))
async function load() {
  emit('clearSelection')
  clearTimeout(durationTimer)
  const current = ++version
  loading.value = true
  error.value = ''
  members.value = []
  groups.value = []
  const snapshot = JSON.parse(JSON.stringify(query.value)) as CategoryQuery
  try {
    if (selected.value) {
      const memberQuery = {
        ...snapshot,
        categoryId: selected.value.categoryId,
        part: all.value ? 0 : selected.value.part,
      }
      const result = await categoryBrowseServer.members(memberQuery)
      if (current !== version) return
      if (!result.status) throw new Error(result.msg)
      const parts =
        snapshot.groupLimit === 0
          ? 1
          : Math.max(1, Math.ceil(result.data.total / snapshot.groupLimit))
      const part = Math.min(selected.value!.part, parts)
      selected.value = {
        ...selected.value!,
        part,
        parts,
        total: result.data.total,
        count:
          snapshot.groupLimit === 0
            ? result.data.total
            : Math.max(
                0,
                Math.min(snapshot.groupLimit, result.data.total - (part - 1) * snapshot.groupLimit),
              ),
      }
      if (!all.value && result.data.total > 0 && memberQuery.part !== part) {
        // 资料减少后末组可能消失；重新读取仍存在的最后一组。
        await load()
        return
      }
      const lastPage = Math.max(1, Math.ceil(result.data.total / pageSize.value))
      if (all.value && page.value > lastPage) {
        page.value = lastPage
        return
      }
      members.value = result.data.dataList
      memberTotal.value = result.data.total
      lastMembersQuery = memberQuery
      scheduleDurations()
    } else {
      const result = await categoryBrowseServer.summaries(snapshot)
      if (current !== version) return
      if (!result.status) throw new Error(result.msg)
      const lastPage = Math.max(1, Math.ceil(result.data.total / pageSize.value))
      if (page.value > lastPage) {
        page.value = lastPage
        return
      }
      groups.value = result.data.dataList
      total.value = result.data.total
      resourceTotal.value = result.data.resourceTotal
      categoryTotal.value = result.data.categoryTotal
    }
  } catch (e) {
    if (current === version) error.value = e instanceof Error ? e.message : '分类浏览加载失败'
  } finally {
    if (current === version) loading.value = false
  }
}
function scheduleDurations() {
  clearTimeout(durationTimer)
  if (
    !store.currentConfigApp.showVideoDuration ||
    !selected.value ||
    !lastMembersQuery ||
    !members.value.length
  )
    return
  const current = version,
    snapshot = lastMembersQuery
  durationTimer = setTimeout(async () => {
    if (current !== version || !store.currentConfigApp.showVideoDuration) return
    try {
      const result = await categoryBrowseServer.members(snapshot)
      if (current !== version || !result.status) return
      const fresh = new Map(result.data.dataList.map((item) => [item.id, item]))
      for (const item of members.value) {
        const durations = new Map(
          (fresh.get(item.id)?.dramaSeries || []).map((series) => [
            series.id,
            series.durationSeconds,
          ]),
        )
        for (const series of item.dramaSeries || []) {
          const duration = durations.get(series.id)
          if (typeof duration === 'number') series.durationSeconds = duration
        }
      }
    } catch {
      /* 补采失败保留当前列表，不改变排序和预览状态。 */
    }
  }, 2500)
}
watch(
  () => store.currentConfigApp.showVideoDuration,
  () => scheduleDurations(),
)
function enter(group: CategoryGroup, full: boolean) {
  if (!selected.value) categoryPage = page.value
  selected.value = group
  all.value = full || settings.value.categoryBrowseGroupLimit === 0
  memberTotal.value = group.total
  reloadAt(1)
}
function back() {
  selected.value = undefined
  all.value = false
  reloadAt(categoryPage)
}
function reloadAt(target: number) {
  if (page.value === target) void load()
  else page.value = target
}
watch(
  page,
  () => {
    void load()
  },
  { flush: 'sync' },
)
watch(
  () =>
    JSON.stringify([
      store.currentFilesBases.id,
      search.searchData,
      settings.value.categoryBrowseSource,
      settings.value.categoryBrowseSort,
      settings.value.categoryBrowseGroupLimit,
      pageSize.value,
    ]),
  () => {
    version++
    selected.value = undefined
    all.value = false
    total.value = 0
    resourceTotal.value = 0
    categoryTotal.value = 0
    memberTotal.value = 0
    categoryPage = 1
    reloadAt(1)
  },
  { immediate: true },
)
onBeforeUnmount(() => {
  version++
  clearTimeout(durationTimer)
})
const pagination = computed(() => ({
  page: page.value,
  pageSize: pageSize.value,
  total: selected.value ? memberTotal.value : total.value,
  visible: !selected.value || all.value,
  shortSummary: selected.value
    ? all.value
      ? `${memberTotal.value} 资源`
      : `本组 ${members.value.length} / 共 ${memberTotal.value}`
    : `${total.value} 组 · ${resourceTotal.value} 资源`,
  summary: selected.value
    ? `共 ${memberTotal.value} 个资源${all.value ? '' : ` · 本组 ${members.value.length} 个`}`
    : `共 ${categoryTotal.value} 个分类 · ${total.value} 个展示组 · ${resourceTotal.value} 个资源`,
}))
function setPage(value: number) {
  page.value = value
}
defineExpose({ refresh: load, showDataList: () => members.value, pagination, setPage })
</script>
<style scoped>
.category-browse {
  color: var(--el-text-color-primary);
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  gap: 12px;
  padding: 10px;
  box-sizing: border-box;
}
.category-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.category-toolbar small {
  display: block;
  margin-top: 4px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.category-toolbar > div:last-child {
  display: flex;
  align-items: center;
  gap: 12px;
}
.category-breadcrumb {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}
.category-breadcrumb > span {
  color: var(--el-text-color-secondary);
}
.category-breadcrumb button {
  border: 0;
  padding: 0;
  background: transparent;
  color: var(--el-color-primary);
  font: inherit;
  cursor: pointer;
}
.category-breadcrumb button:focus-visible {
  outline: 2px solid var(--el-color-primary);
  outline-offset: 3px;
}
.category-scroll,
.category-resources {
  flex: 1;
  min-height: 0;
}
.category-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
  padding: 4px;
}
footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  flex-wrap: wrap;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.category-error {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}
</style>
