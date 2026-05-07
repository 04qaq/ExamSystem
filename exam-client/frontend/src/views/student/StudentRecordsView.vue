<script setup lang="ts">
import { NCard, NDataTable, NButton, NPagination, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { h, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import type { ExamRecordRow } from '@/api/types'
import { studentRecords } from '@/api/student'

const message = useMessage()
const router = useRouter()

const loading = ref(false)
const items = ref<ExamRecordRow[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)

const statusMap: Record<number, { label: string; type: 'default' | 'info' | 'success' | 'warning' }> = {
  1: { label: '进行中', type: 'warning' },
  2: { label: '已提交', type: 'info' },
  3: { label: '已批阅', type: 'success' },
}

const columns: DataTableColumns<ExamRecordRow> = [
  { title: '试卷', key: 'paper_title' },
  { title: '得分', key: 'total_score', width: 80 },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render(row) {
      const s = statusMap[row.status] || { label: String(row.status), type: 'default' }
      return h('span', {}, s.label)
    },
  },
  { title: '开始时间', key: 'start_time', width: 170 },
  {
    title: '交卷时间',
    key: 'submit_time',
    width: 170,
    render(row) {
      return row.submit_time || '—'
    },
  },
  {
    title: '操作',
    key: 'actions',
    width: 100,
    render(row) {
      return h(
        NButton,
        {
          size: 'small',
          type: row.status === 1 ? 'warning' : 'primary',
          onClick: () =>
            row.status === 1
              ? router.push({ name: 'student-exam', params: { paperId: String(row.paper_id) } })
              : router.push({ name: 'student-record-detail', params: { id: String(row.id) } }),
        },
        { default: () => (row.status === 1 ? '继续' : '详情') }
      )
    },
  },
]

async function load() {
  loading.value = true
  try {
    const res = await studentRecords(page.value, pageSize.value)
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
  <n-card title="我的成绩">
    <n-data-table :columns="columns" :data="items" :loading="loading" :bordered="false" />
    <div style="margin-top: 16px; display: flex; justify-content: flex-end">
      <n-pagination
        v-model:page="page"
        v-model:page-size="pageSize"
        :item-count="total"
        show-size-picker
        :page-sizes="[10, 20, 50]"
        @update:page="load"
        @update:page-size="load"
      />
    </div>
  </n-card>
</template>
