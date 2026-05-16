<script setup lang="ts">
import {
  NCard,
  NButton,
  NSpace,
  NRadioGroup,
  NRadio,
  NCheckboxGroup,
  NCheckbox,
  NInput,
  NTag,
  useMessage,
  useDialog,
} from 'naive-ui'
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router'
import type { ExamQuestionItem } from '@/api/types'
import { recordDetail, saveAnswer, startExam, submitExam } from '@/api/student'
import { parseOptionList } from '@/utils/options'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const dialog = useDialog()

const paperId = computed(() => Number(route.params.paperId))

const loading = ref(true)
const recordId = ref(0)
const questions = ref<ExamQuestionItem[]>([])
const remaining = ref(0)
const totalScore = ref(0)

/** question_id -> answer string */
const answers = ref<Record<number, string>>({})
/** multi select buffer */
const multi = ref<Record<number, string[]>>({})

const tabSwitchCount = ref(0)
const submitted = ref(false)
const saveStatus = ref<'idle' | 'saving' | 'saved' | 'error'>('idle')
const lastSavedAt = ref('')

/** 与后端 exam 记录状态一致：进行中才可回填答案 */
const EXAM_IN_PROGRESS = 1
let saveTimer: ReturnType<typeof setTimeout> | undefined
let tickTimer: ReturnType<typeof setInterval> | undefined
const dirtyQuestionIds = new Set<number>()

function scheduleSave(qid: number) {
  dirtyQuestionIds.add(qid)
  saveStatus.value = 'saving'
  clearTimeout(saveTimer)
  saveTimer = setTimeout(() => {
    void flushPendingAnswers()
  }, 700)
}

async function flushPendingAnswers() {
  if (!recordId.value || dirtyQuestionIds.size === 0) return
  const ids = [...dirtyQuestionIds]
  dirtyQuestionIds.clear()
  saveStatus.value = 'saving'
  try {
    await Promise.all(ids.map((qid) => saveAnswer(recordId.value, qid, answers.value[qid] ?? '')))
    saveStatus.value = 'saved'
    lastSavedAt.value = new Date().toLocaleTimeString()
  } catch {
    ids.forEach((qid) => dirtyQuestionIds.add(qid))
    saveStatus.value = 'error'
  }
}

function setSingle(q: ExamQuestionItem, val: string) {
  answers.value[q.question_id] = val
  scheduleSave(q.question_id)
}

function updateMulti(q: ExamQuestionItem, vals: string[]) {
  multi.value[q.question_id] = vals
  const sorted = [...vals].sort()
  answers.value[q.question_id] = sorted.join(',')
  scheduleSave(q.question_id)
}

function setFillShort(q: ExamQuestionItem, val: string) {
  answers.value[q.question_id] = val
  scheduleSave(q.question_id)
}

function judgeOptions(q: ExamQuestionItem) {
  const o = parseOptionList(q.options)
  return o.length >= 2 ? o : ['正确', '错误']
}

function typeLabel(t: number) {
  const m: Record<number, string> = {
    1: '单选题',
    2: '多选题',
    3: '判断题',
    4: '填空题',
    5: '简答题',
  }
  return m[t] || '题目'
}

async function load() {
  submitted.value = false
  loading.value = true
  answers.value = {}
  try {
    const res = await startExam(paperId.value)
    recordId.value = res.record_id
    questions.value = [...res.questions].sort((a, b) => a.sort_order - b.sort_order)
    totalScore.value = res.total_score
    remaining.value = res.remaining_seconds
    const initMulti: Record<number, string[]> = {}
    for (const q of questions.value) {
      if (q.type === 2) initMulti[q.question_id] = []
    }
    multi.value = initMulti

    try {
      const detail = await recordDetail(res.record_id)
      if (detail.status === EXAM_IN_PROGRESS && detail.details?.length) {
        const mergedAns: Record<number, string> = {}
        const mergedMulti = { ...initMulti }
        for (const d of detail.details) {
          const pa = (d.provided_answer ?? '').trim()
          if (!pa) continue
          if (d.type === 2) {
            const parts = pa.split(',').map((s) => s.trim()).filter(Boolean)
            mergedMulti[d.question_id] = parts
            mergedAns[d.question_id] = [...parts].sort((a, b) => a.localeCompare(b)).join(',')
          } else {
            mergedAns[d.question_id] = d.provided_answer ?? ''
          }
        }
        answers.value = mergedAns
        multi.value = mergedMulti
      }
    } catch {
      /* 详情失败不影响新开考 */
    }
  } catch (e) {
    message.error(e instanceof Error ? e.message : '无法开始考试')
    router.back()
  } finally {
    loading.value = false
  }
}

function formatRemain(sec: number) {
  const m = Math.floor(sec / 60)
  const s = sec % 60
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

function isAnswered(q: ExamQuestionItem) {
  return !!(answers.value[q.question_id] ?? '').trim()
}

const answeredCount = computed(() => questions.value.filter(isAnswered).length)
const unansweredCount = computed(() => Math.max(0, questions.value.length - answeredCount.value))

const saveStatusText = computed(() => {
  if (saveStatus.value === 'saving') return '保存中'
  if (saveStatus.value === 'saved') return lastSavedAt.value ? `已保存 ${lastSavedAt.value}` : '已保存'
  if (saveStatus.value === 'error') return '保存失败，请检查网络'
  return '尚未作答'
})

const saveStatusType = computed(() => {
  if (saveStatus.value === 'error') return 'error'
  if (saveStatus.value === 'saving') return 'warning'
  if (saveStatus.value === 'saved') return 'success'
  return 'default'
})

function scrollToQuestion(qid: number) {
  document.getElementById(`q-${qid}`)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

function onVisibility() {
  if (document.visibilityState === 'hidden') tabSwitchCount.value += 1
}

async function doSubmit() {
  if (submitted.value) return
  submitted.value = true
  try {
    clearTimeout(saveTimer)
    await flushPendingAnswers()
    if (saveStatus.value === 'error') {
      submitted.value = false
      message.error('仍有答案保存失败，请网络恢复后再交卷')
      return
    }
    const res = await submitExam(recordId.value, tabSwitchCount.value)
    message.success(`已提交，得分 ${res.total_score}（客观题已自动阅卷）`)
    router.replace({ name: 'student-record-detail', params: { id: String(recordId.value) } })
  } catch (e) {
    submitted.value = false
    message.error(e instanceof Error ? e.message : '提交失败')
  }
}

function confirmSubmit() {
  dialog.warning({
    title: '提交试卷',
    content: unansweredCount.value > 0
      ? `还有 ${unansweredCount.value} 题未作答。提交后不可修改，确定提交吗？`
      : '所有题目均已作答。提交后不可修改，确定提交吗？',
    positiveText: '提交',
    negativeText: '再检查一下',
    onPositiveClick: () => {
      void doSubmit()
    },
  })
}

watch(paperId, () => {
  void load()
})

onMounted(() => {
  void load()
  document.addEventListener('visibilitychange', onVisibility)
  tickTimer = setInterval(() => {
    if (remaining.value <= 0 || submitted.value) return
    remaining.value -= 1
    if (remaining.value === 0 && !submitted.value) {
      message.warning('考试时间已到，系统将尝试自动提交')
      void doSubmit()
    }
  }, 1000)
})

onBeforeUnmount(() => {
  document.removeEventListener('visibilitychange', onVisibility)
  clearInterval(tickTimer)
  clearTimeout(saveTimer)
})

let allowLeave = false
onBeforeRouteLeave((_to, _from, next) => {
  if (allowLeave) {
    next()
    return
  }
  dialog.warning({
    title: '暂离考试',
    content: saveStatus.value === 'error'
      ? '当前有答案保存失败，离开可能导致最近作答丢失。确定仍要离开吗？'
      : '答题进度会自动保存，建议确认保存状态为“已保存”后再离开。',
    positiveText: '继续答题',
    negativeText: '仍要离开',
    onNegativeClick: () => {
      allowLeave = true
      next()
    },
    onPositiveClick: () => next(false),
  })
})

</script>

<template>
  <div>
    <n-card size="small" style="margin-bottom: 16px; position: sticky; top: 0; z-index: 5">
      <n-space justify="space-between" align="center" style="flex-wrap: wrap">
        <n-space>
          <strong>考试中</strong>
          <n-tag :type="remaining < 300 ? 'error' : 'info'" size="large">
            剩余 {{ formatRemain(remaining) }}
          </n-tag>
          <n-tag :type="saveStatusType">{{ saveStatusText }}</n-tag>
          <span class="muted">已答 {{ answeredCount }} / {{ questions.length }}</span>
          <span class="muted">切屏 {{ tabSwitchCount }} 次（切到其他窗口或最小化可能被记录，仅供监考参考）</span>
        </n-space>
        <n-space>
          <n-button @click="router.back()">暂离考试</n-button>
          <n-button type="primary" @click="confirmSubmit">交卷</n-button>
        </n-space>
      </n-space>
      <n-space class="question-nav" size="small">
        <n-button
          v-for="q in questions"
          :key="q.question_id"
          size="tiny"
          :type="isAnswered(q) ? 'primary' : 'default'"
          secondary
          @click="scrollToQuestion(q.question_id)"
        >
          {{ q.sort_order }}
        </n-button>
      </n-space>
    </n-card>

    <n-spin :show="loading">
      <n-space vertical size="large">
        <n-card
          v-for="q in questions"
          :id="`q-${q.question_id}`"
          :key="q.question_id"
          :title="`${q.sort_order}. [${typeLabel(q.type)}] (${q.score} 分)`"
        >
          <div class="stem">{{ q.content }}</div>

          <template v-if="q.type === 1">
            <n-radio-group
              :value="answers[q.question_id] ?? null"
              @update:value="(v) => setSingle(q, String(v))"
            >
              <n-space vertical>
                <n-radio v-for="(opt, idx) in parseOptionList(q.options)" :key="idx" :value="opt">
                  {{ opt }}
                </n-radio>
              </n-space>
            </n-radio-group>
          </template>

          <template v-else-if="q.type === 2">
            <n-checkbox-group
              :value="multi[q.question_id] ?? []"
              @update:value="(vals: string[]) => updateMulti(q, vals)"
            >
              <n-space vertical>
                <n-checkbox v-for="(opt, idx) in parseOptionList(q.options)" :key="idx" :value="opt" :label="opt" />
              </n-space>
            </n-checkbox-group>
          </template>

          <template v-else-if="q.type === 3">
            <n-radio-group
              :value="answers[q.question_id] ?? null"
              @update:value="(v) => setSingle(q, String(v))"
            >
              <n-space>
                <n-radio v-for="(opt, idx) in judgeOptions(q)" :key="idx" :value="opt">
                  {{ opt }}
                </n-radio>
              </n-space>
            </n-radio-group>
          </template>

          <template v-else-if="q.type === 4">
            <n-input
              :value="answers[q.question_id] ?? ''"
              placeholder="请填写答案"
              @update:value="(v) => setFillShort(q, v)"
            />
          </template>

          <template v-else-if="q.type === 5">
            <n-input
              type="textarea"
              :rows="5"
              :value="answers[q.question_id] ?? ''"
              placeholder="请作答"
              @update:value="(v) => setFillShort(q, v)"
            />
          </template>
        </n-card>
      </n-space>
    </n-spin>
  </div>
</template>

<style scoped>
.stem {
  margin-bottom: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
}
.muted {
  font-size: 13px;
  color: #64748b;
}
.question-nav {
  margin-top: 12px;
  max-height: 88px;
  overflow: auto;
}
</style>
