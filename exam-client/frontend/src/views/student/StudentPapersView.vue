<script setup lang="ts">
import { NCard, NButton, NSpace, NTag, NSpin, NEmpty, useMessage } from 'naive-ui'
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import type { AvailablePaper, ExamRecordRow } from '@/api/types'
import { studentPapers, studentRecords } from '@/api/student'

const message = useMessage()
const router = useRouter()

const loading = ref(true)
const papers = ref<AvailablePaper[]>([])
const latestRecordByPaper = ref<Map<number, ExamRecordRow>>(new Map())

const ExamInProgress = 1
const ExamSubmitted = 2
const ExamGraded = 3

async function load() {
  loading.value = true
  try {
    const [plist, recData] = await Promise.all([
      studentPapers(),
      studentRecords(1, 500),
    ])
    papers.value = plist
    const m = new Map<number, ExamRecordRow>()
    for (const r of recData.items) {
      const prev = m.get(r.paper_id)
      if (!prev || r.id > prev.id) m.set(r.paper_id, r)
    }
    latestRecordByPaper.value = m
  } catch (e) {
    message.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(load)

function statusForPaper(p: AvailablePaper) {
  const rec = latestRecordByPaper.value.get(p.id)
  if (!rec) return 'fresh' as const
  if (rec.status === ExamInProgress) return 'progress' as const
  if (rec.status === ExamSubmitted || rec.status === ExamGraded) return 'done' as const
  return 'fresh' as const
}

const rows = computed(() =>
  papers.value.map((p) => ({
    paper: p,
    status: statusForPaper(p),
  }))
)

function goExam(id: number) {
  router.push({ name: 'student-exam', params: { paperId: String(id) } })
}

function goResult(recordId: number) {
  router.push({ name: 'student-record-detail', params: { id: String(recordId) } })
}
</script>

<template>
  <n-spin :show="loading">
    <n-space vertical size="large">
      <n-card title="可考试卷" size="small">
        <template #header-extra>
          <n-button size="small" @click="load">刷新</n-button>
        </template>
        <n-empty v-if="!loading && rows.length === 0" description="当前没有可选试卷（请关注考试时间窗口）" />
        <n-card
          v-for="row in rows"
          :key="row.paper.id"
          size="small"
          class="paper-card"
          :title="row.paper.title"
        >
          <p class="desc">{{ row.paper.description || '暂无说明' }}</p>
          <n-space align="center" style="flex-wrap: wrap">
            <span>总分 {{ row.paper.total_score }} · 时长 {{ row.paper.duration }} 分钟</span>
            <n-tag size="small" type="info">{{ row.paper.start_time }} ~ {{ row.paper.end_time }}</n-tag>
            <n-tag v-if="row.status === 'progress'" type="warning">进行中</n-tag>
            <n-tag v-else-if="row.status === 'done'" type="success">已提交</n-tag>
            <n-tag v-else type="default">未开始</n-tag>
          </n-space>
          <n-space style="margin-top: 12px">
            <template v-if="row.status === 'done'">
              <n-button type="primary" @click="goResult(latestRecordByPaper.get(row.paper.id)!.id)">
                查看成绩
              </n-button>
            </template>
            <template v-else>
              <n-button type="primary" @click="goExam(row.paper.id)">
                {{ row.status === 'progress' ? '继续答题' : '进入考试' }}
              </n-button>
            </template>
          </n-space>
        </n-card>
      </n-card>
    </n-space>
  </n-spin>
</template>

<style scoped>
.paper-card + .paper-card {
  margin-top: 12px;
}
.desc {
  color: #64748b;
  margin: 0 0 8px;
  font-size: 14px;
}
</style>
