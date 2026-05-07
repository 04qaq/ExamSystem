<script setup lang="ts">
import { NCard, NDataTable, NPagination, NSpace, NInput, NButton, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { onMounted, ref } from 'vue'
import type { LogRow } from '@/api/admin'
import { adminLogs } from '@/api/admin'

const message = useMessage()

const loading = ref(false)
const items = ref<LogRow[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(15)
const action = ref('')

const columns: DataTableColumns<LogRow> = [
  { title: '时间', key: 'created_at', width: 170 },
  { title: '用户', key: 'username', width: 120 },
  { title: '动作', key: 'action', width: 120 },
  { title: '对象', key: 'target', ellipsis: { tooltip: true } },
  { title: 'IP', key: 'ip', width: 130 },
]

async function load() {
  loading.value = true
  try {
    const res = await adminLogs(page.value, pageSize.value, action.value || undefined)
    items.value = res.items
    total.value = res.total
  } catch (e) {
    message.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <n-card title="操作日志">
    <n-space style="margin-bottom: 16px">
      <n-input v-model:value="action" placeholder="按动作筛选" clearable style="width: 200px" @keyup.enter="load" />
      <n-button type="primary" @click="((page = 1), load())">查询</n-button>
    </n-space>
    <n-data-table :columns="columns" :data="items" :loading="loading" />
    <div style="margin-top: 16px; display: flex; justify-content: flex-end">
      <n-pagination
        v-model:page="page"
        v-model:page-size="pageSize"
        :item-count="total"
        show-size-picker
        :page-sizes="[15, 30, 50]"
        @update:page="load"
        @update:page-size="load"
      />
    </div>
  </n-card>
</template>
