<script setup lang="ts">
import {
  NCard,
  NDataTable,
  NPagination,
  NModal,
  NSpin,
  NSpace,
  NButton,
  NInputNumber,
  NInput,
  useMessage,
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { h, onMounted, reactive, ref } from 'vue'
import type { PendingRow, GradeDetailItem } from '@/api/teacher'
import { pendingGrades, gradeDetail, gradeSubmit } from '@/api/teacher'
import { parseOptionList } from '@/utils/options'

const message = useMessage()

const loading = ref(false)
const items = ref<PendingRow[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)

const show = ref(false)
const detailLoading = ref(false)
const currentRecordId = ref(0)
const detailMeta = reactive({
  paper_title: '',
  student_name: '',
})
const detailRows = ref<GradeDetailItem[]>([])
const grades = reactive<Record<number, { score: number; comment: string }>>({})

const columns: DataTableColumns<PendingRow> = [
  { title: '试卷', key: 'paper_title', ellipsis: { tooltip: true } },
  { title: '学生', key: 'student_name', width: 120 },
  { title: '提交时间', key: 'submit_time', width: 170 },
  {
    title: '主观进度',
    key: 'pg',
    width: 120,
    render(row) {
      return `${row.graded_count} / ${row.subjective_count}`
    },
  },
  {
    title: '操作',
    key: 'act',
    width: 100,
    render(row) {
      return h(
        NButton,
        { size: 'small', type: 'primary', onClick: () => openGrade(row) },
        { default: () => '批阅' }
      )
    },
  },
]

async function load() {
  loading.value = true
  try {
    const res = await pendingGrades(page.value, pageSize.value)
    items.value = res.items
    total.value = res.total
  } catch (e) {
    message.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

function needsGrade(t: number) {
  return t === 4 || t === 5
}

async function openGrade(row: PendingRow) {
  currentRecordId.value = row.record_id
  show.value = true
  detailLoading.value = true
  detailRows.value = []
  for (const k of Object.keys(grades)) {
    delete grades[Number(k)]
  }
  try {
    const d = await gradeDetail(row.record_id)
    detailMeta.paper_title = d.paper_title
    detailMeta.student_name = d.student_name
    detailRows.value = d.details
    for (const r of d.details) {
      if (needsGrade(r.type)) {
        grades[r.detail_id] = {
          score: r.score_gained,
          comment: r.comment || '',
        }
      }
    }
  } catch (e) {
    message.error(e instanceof Error ? e.message : '加载详情失败')
    show.value = false
  } finally {
    detailLoading.value = false
  }
}

async function submitGrades() {
  const list: { detail_id: number; score_gained: number; comment: string }[] = []
  for (const r of detailRows.value) {
    if (!needsGrade(r.type)) continue
    const g = grades[r.detail_id]
    if (!g) continue
    list.push({
      detail_id: r.detail_id,
      score_gained: Math.min(Math.max(0, g.score), r.score),
      comment: g.comment || '',
    })
  }
  try {
    await gradeSubmit(currentRecordId.value, list)
    message.success('批阅已提交')
    show.value = false
    await load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '提交失败')
  }
}

onMounted(load)
</script>

<template>
  <n-card title="待批阅">
    <n-data-table :columns="columns" :data="items" :loading="loading" />
    <div style="margin-top: 16px; display: flex; justify-content: flex-end">
      <n-pagination
        v-model:page="page"
        v-model:page-size="pageSize"
        :item-count="total"
        show-size-picker
        :page-sizes="[10, 20]"
        @update:page="load"
        @update:page-size="load"
      />
    </div>
  </n-card>

  <n-modal
    v-model:show="show"
    preset="card"
    :title="`${detailMeta.paper_title} — ${detailMeta.student_name}`"
    style="width: min(960px, 96vw)"
    :mask-closable="false"
  >
    <n-spin :show="detailLoading">
      <div v-for="r in detailRows" :key="r.detail_id" class="block">
        <p class="stem">{{ r.content }}</p>
        <ul v-if="parseOptionList(r.options).length" class="opts">
          <li v-for="(o, i) in parseOptionList(r.options)" :key="i">{{ o }}</li>
        </ul>
        <p><span class="muted">学生作答：</span>{{ r.provided_answer || '（空）' }}</p>
        <p><span class="muted">参考答案：</span>{{ r.correct_answer }}</p>
        <template v-if="needsGrade(r.type)">
          <n-space align="center" style="margin-top: 8px">
            <span>得分</span>
            <n-input-number
              v-model:value="grades[r.detail_id].score"
              :min="0"
              :max="r.score"
              size="small"
            />
            <span>/ {{ r.score }}</span>
          </n-space>
          <n-input
            v-model:value="grades[r.detail_id].comment"
            type="textarea"
            placeholder="评语（可选）"
            :rows="2"
            style="margin-top: 8px"
          />
        </template>
        <template v-else>
          <p class="muted">客观题由系统自动阅卷 · 得分 {{ r.score_gained }} / {{ r.score }}</p>
        </template>
      </div>
      <n-space justify="end" style="margin-top: 16px">
        <n-button @click="show = false">取消</n-button>
        <n-button type="primary" @click="submitGrades">提交批阅</n-button>
      </n-space>
    </n-spin>
  </n-modal>
</template>

<style scoped>
.block {
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
}
.stem {
  white-space: pre-wrap;
  line-height: 1.55;
}
.opts {
  margin: 4px 0 8px;
  padding-left: 18px;
}
.muted {
  color: #64748b;
  font-size: 13px;
}
</style>
