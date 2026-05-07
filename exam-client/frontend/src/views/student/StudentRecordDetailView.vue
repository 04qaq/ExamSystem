<script setup lang="ts">
import { NCard, NTag, NSpace, NSpin, useMessage } from 'naive-ui'
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { recordDetail } from '@/api/student'
import type { AnswerDetailRow } from '@/api/student'
import { parseOptionList } from '@/utils/options'

const route = useRoute()
const message = useMessage()

const loading = ref(true)
const detail = ref<Awaited<ReturnType<typeof recordDetail>> | null>(null)

const id = computed(() => Number(route.params.id))

function verdict(row: AnswerDetailRow) {
  if (row.type === 5 || row.type === 4) {
    if (row.is_correct === null) return { text: '待批阅', type: 'warning' as const }
  }
  if (row.is_correct === 1) return { text: '正确', type: 'success' as const }
  if (row.is_correct === 0) return { text: '错误', type: 'error' as const }
  return { text: '—', type: 'default' as const }
}

onMounted(async () => {
  try {
    detail.value = await recordDetail(id.value)
  } catch (e) {
    message.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <n-spin :show="loading">
    <template v-if="detail">
      <n-card :title="detail.paper_title" style="margin-bottom: 16px">
        <n-space>
          <span>总分：<strong>{{ detail.total_score }}</strong></span>
          <n-tag size="small">开始 {{ detail.start_time }}</n-tag>
          <n-tag v-if="detail.submit_time" size="small" type="info">交卷 {{ detail.submit_time }}</n-tag>
        </n-space>
      </n-card>
      <n-card v-for="(row, idx) in detail.details" :key="idx" :title="`第 ${idx + 1} 题`" size="small" style="margin-bottom: 12px">
        <p class="stem">{{ row.content }}</p>
        <template v-if="parseOptionList(row.options).length">
          <p class="label">选项</p>
          <ul class="opts">
            <li v-for="(o, i) in parseOptionList(row.options)" :key="i">{{ o }}</li>
          </ul>
        </template>
        <p><span class="label">你的答案：</span>{{ row.provided_answer || '（未作答）' }}</p>
        <p><span class="label">参考答案：</span>{{ row.correct_answer }}</p>
        <n-space align="center">
          <span class="label">判定</span>
          <n-tag :type="verdict(row).type">{{ verdict(row).text }}</n-tag>
          <span>得分 {{ row.score_gained }} / {{ row.score }}</span>
        </n-space>
        <p v-if="row.comment" class="comment">评语：{{ row.comment }}</p>
      </n-card>
    </template>
  </n-spin>
</template>

<style scoped>
.stem {
  white-space: pre-wrap;
  line-height: 1.6;
}
.label {
  color: #64748b;
  font-size: 13px;
}
.opts {
  margin: 4px 0 12px;
  padding-left: 20px;
}
.comment {
  margin-top: 8px;
  color: #475569;
}
</style>
