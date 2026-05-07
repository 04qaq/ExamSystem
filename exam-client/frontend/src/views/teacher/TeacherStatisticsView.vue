<script setup lang="ts">
import { NCard, NSpin, NSpace, NButton, NStatistic, NGrid, NGi, useMessage } from 'naive-ui'
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { paperStatistics } from '@/api/teacher'
import { getBlob } from '@/api/http'
import { downloadBlob } from '@/utils/download'

const route = useRoute()
const message = useMessage()

const paperId = computed(() => Number(route.params.paperId))
const loading = ref(true)
const stats = ref<Awaited<ReturnType<typeof paperStatistics>> | null>(null)

async function load() {
  loading.value = true
  try {
    stats.value = await paperStatistics(paperId.value)
  } catch (e) {
    message.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

async function exportExcel() {
  try {
    const { blob, filename } = await getBlob(`/api/teacher/statistics/paper/${paperId.value}/export`)
    downloadBlob(blob, filename)
    message.success('已开始下载')
  } catch (e) {
    message.error(e instanceof Error ? e.message : '导出失败')
  }
}

onMounted(load)
</script>

<template>
  <n-spin :show="loading">
    <template v-if="stats">
      <n-card :title="`统计 · ${stats.paper.paper_title}`">
        <n-space style="margin-bottom: 20px">
          <n-button type="primary" @click="exportExcel">导出成绩 Excel</n-button>
          <n-button @click="load">刷新</n-button>
        </n-space>
        <n-grid :cols="2" :x-gap="16" :y-gap="16" responsive="screen">
          <n-gi>
            <n-statistic label="参考人数" :value="stats.paper.total_students" />
          </n-gi>
          <n-gi>
            <n-statistic label="已交卷" :value="stats.paper.submitted_count" />
          </n-gi>
          <n-gi>
            <n-statistic label="平均分" :value="stats.paper.avg_score" :precision="2" />
          </n-gi>
          <n-gi>
            <n-statistic label="最高分" :value="stats.paper.max_score" />
          </n-gi>
          <n-gi>
            <n-statistic label="最低分" :value="stats.paper.min_score" />
          </n-gi>
          <n-gi>
            <n-statistic label="及格率" :value="stats.paper.pass_rate * 100" suffix="%" :precision="1" />
          </n-gi>
        </n-grid>
      </n-card>
      <n-card title="题目正确率" style="margin-top: 16px">
        <table class="tbl">
          <thead>
            <tr>
              <th>题干摘要</th>
              <th>正确率</th>
              <th>平均分</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="q in stats.question_stats" :key="q.question_id">
              <td>{{ q.content.slice(0, 80) }}{{ q.content.length > 80 ? '…' : '' }}</td>
              <td>{{ (q.correct_rate * 100).toFixed(1) }}%</td>
              <td>{{ q.avg_score.toFixed(2) }}</td>
            </tr>
          </tbody>
        </table>
      </n-card>
    </template>
  </n-spin>
</template>

<style scoped>
.tbl {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}
.tbl th,
.tbl td {
  border: 1px solid rgba(0, 0, 0, 0.08);
  padding: 8px 10px;
  text-align: left;
}
</style>
