<script setup lang="ts">
import {
  NCard,
  NDataTable,
  NPagination,
  NSpace,
  NButton,
  NInput,
  NSelect,
  NModal,
  NTag,
  NDropdown,
  NForm,
  NFormItem,
  NInputNumber,
  NDatePicker,
  NSpin,
  useMessage,
  useDialog,
  type DropdownOption,
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { h, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import type { PaperRow, QuestionRow } from '@/api/types'
import {
  paperPreview,
  teacherPaperAddQuestions,
  teacherPaperCopy,
  teacherPaperCreate,
  teacherPaperDelete,
  teacherPaperGet,
  teacherPaperList,
  teacherPaperPublish,
  teacherPaperRandomSelect,
  teacherPaperUnpublish,
  teacherPaperUpdate,
  teacherQuestionList,
  type PaperSavePayload,
} from '@/api/teacher'
import { parseBackendDateTime, toApiDateTime } from '@/utils/time'
import { parseOptionList } from '@/utils/options'

const message = useMessage()
const dialog = useDialog()
const router = useRouter()

const loading = ref(false)
const items = ref<PaperRow[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const keyword = ref('')
const status = ref<number | undefined>(undefined)

const statusOpts = [
  { label: '草稿', value: 0 },
  { label: '已发布', value: 1 },
  { label: '已关闭', value: 2 },
]

const statusMap: Record<number, { label: string; type: 'default' | 'success' | 'warning' | 'error' }> = {
  0: { label: '草稿', type: 'default' },
  1: { label: '已发布', type: 'success' },
  2: { label: '已关闭', type: 'warning' },
}

const showPreview = ref(false)
const previewLoading = ref(false)
const previewTitle = ref('')
const previewQs = ref<{ sort: number; content: string; type?: number; score: number; options?: string }[]>([])

async function openPreview(row: PaperRow) {
  showPreview.value = true
  previewLoading.value = true
  previewTitle.value = row.title
  previewQs.value = []
  try {
    const res = await paperPreview(row.id)
    previewQs.value = res.questions.map((q) => ({
      sort: q.sort_order,
      content: q.question_content || '',
      type: q.question_type,
      score: q.score,
      options: q.question_options,
    }))
  } catch (e) {
    message.error(e instanceof Error ? e.message : '预览失败')
    showPreview.value = false
  } finally {
    previewLoading.value = false
  }
}

/** ---------- 试卷基本信息 ---------- */
const showPaperModal = ref(false)
const paperSaving = ref(false)
const editingPaperId = ref<number | null>(null)
const paperForm = ref({
  title: '',
  description: '',
  duration: 90,
  total_score: 100,
  startTs: Date.now(),
  endTs: Date.now() + 7 * 24 * 3600 * 1000,
})

function openCreatePaper() {
  editingPaperId.value = null
  const now = Date.now()
  paperForm.value = {
    title: '',
    description: '',
    duration: 90,
    total_score: 100,
    startTs: now,
    endTs: now + 7 * 24 * 3600 * 1000,
  }
  showPaperModal.value = true
}

async function openEditPaper(row: PaperRow) {
  if (row.status !== 0) {
    message.warning('仅草稿可编辑基本信息')
    return
  }
  editingPaperId.value = row.id
  try {
    const p = await teacherPaperGet(row.id)
    paperForm.value = {
      title: p.title,
      description: p.description || '',
      duration: p.duration,
      total_score: p.total_score,
      startTs: parseBackendDateTime(p.start_time),
      endTs: parseBackendDateTime(p.end_time),
    }
    showPaperModal.value = true
  } catch (e) {
    message.error(e instanceof Error ? e.message : '加载试卷失败')
  }
}

function paperPayload(): PaperSavePayload {
  const f = paperForm.value
  return {
    title: f.title.trim(),
    description: f.description.trim(),
    duration: f.duration,
    total_score: f.total_score,
    start_time: toApiDateTime(f.startTs),
    end_time: toApiDateTime(f.endTs),
  }
}

async function submitPaperForm() {
  if (!paperForm.value.title.trim()) {
    message.warning('请填写标题')
    return
  }
  paperSaving.value = true
  try {
    if (editingPaperId.value != null) {
      await teacherPaperUpdate(editingPaperId.value, paperPayload())
      message.success('已保存')
    } else {
      await teacherPaperCreate(paperPayload())
      message.success('已创建草稿')
    }
    showPaperModal.value = false
    await load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '保存失败')
  } finally {
    paperSaving.value = false
  }
}

/** ---------- 手动选题 ---------- */
const showPickModal = ref(false)
const pickPaperId = ref(0)
const pickLoading = ref(false)
const pickItems = ref<QuestionRow[]>([])
const pickTotal = ref(0)
const pickPage = ref(1)
const pickPageSize = ref(10)
const pickKeyword = ref('')
const pickChecked = ref<number[]>([])
const pickScores = ref<Record<number, number>>({})

watch(pickChecked, (keys) => {
  const m = { ...pickScores.value }
  for (const id of keys) {
    if (m[id] == null) {
      const row = pickItems.value.find((r) => r.id === id)
      if (row) m[id] = row.score
    }
  }
  pickScores.value = m
})

async function loadPickBank() {
  pickLoading.value = true
  try {
    const res = await teacherQuestionList(pickPage.value, pickPageSize.value, {
      keyword: pickKeyword.value || undefined,
    })
    pickItems.value = res.items
    pickTotal.value = res.total
  } catch (e) {
    message.error(e instanceof Error ? e.message : '加载题库失败')
  } finally {
    pickLoading.value = false
  }
}

function openPickModal(row: PaperRow) {
  if (row.status !== 0) {
    message.warning('仅草稿可组卷')
    return
  }
  pickPaperId.value = row.id
  pickChecked.value = []
  pickScores.value = {}
  pickKeyword.value = ''
  pickPage.value = 1
  showPickModal.value = true
  void loadPickBank()
}

async function submitPickQuestions() {
  if (!pickChecked.value.length) {
    message.warning('请勾选题目')
    return
  }
  const questions = pickChecked.value.map((id) => ({
    question_id: id,
    score: pickScores.value[id] ?? 5,
  }))
  try {
    await teacherPaperAddQuestions(pickPaperId.value, questions)
    message.success('已添加到试卷')
    showPickModal.value = false
    await load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '添加失败')
  }
}

const pickColumns: DataTableColumns<QuestionRow> = [
  { type: 'selection' },
  { title: 'ID', key: 'id', width: 70 },
  {
    title: '题型',
    key: 'type',
    width: 64,
    render(row) {
      const m: Record<number, string> = { 1: '单选', 2: '多选', 3: '判断', 4: '填空', 5: '简答' }
      return m[row.type] || row.type
    },
  },
  { title: '题库分值', key: 'score', width: 88 },
  {
    title: '卷内分值',
    key: 'pscore',
    width: 110,
    render(row) {
      return h(NInputNumber, {
        size: 'small',
        min: 1,
        value: pickScores.value[row.id] ?? row.score,
        style: { width: '88px' },
        onUpdateValue: (v: number | null) => {
          pickScores.value = { ...pickScores.value, [row.id]: v ?? row.score }
        },
      })
    },
  },
  {
    title: '题干',
    key: 'content',
    ellipsis: { tooltip: true },
  },
]

/** ---------- 随机组卷 ---------- */
const showRandomModal = ref(false)
const randomPaperId = ref(0)
const randomRules = ref<{ type: number; count: number; total_score: number }[]>([
  { type: 1, count: 5, total_score: 25 },
])

const typeOptsPick = [
  { label: '单选', value: 1 },
  { label: '多选', value: 2 },
  { label: '判断', value: 3 },
  { label: '填空', value: 4 },
  { label: '简答', value: 5 },
]

function openRandomModal(row: PaperRow) {
  if (row.status !== 0) {
    message.warning('仅草稿可随机组卷')
    return
  }
  randomPaperId.value = row.id
  randomRules.value = [{ type: 1, count: 5, total_score: 25 }]
  showRandomModal.value = true
}

function addRandomRule() {
  randomRules.value = [...randomRules.value, { type: 1, count: 1, total_score: 5 }]
}

function removeRandomRule(idx: number) {
  randomRules.value = randomRules.value.filter((_, i) => i !== idx)
}

async function submitRandom() {
  const rules = randomRules.value.filter((r) => r.count >= 1 && r.total_score >= 1)
  if (!rules.length) {
    message.warning('请填写有效规则')
    return
  }
  try {
    await teacherPaperRandomSelect(randomPaperId.value, rules)
    message.success('随机组卷完成')
    showRandomModal.value = false
    await load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '随机组卷失败')
  }
}

/** ---------- 行操作 ---------- */
function paperDropdownOptions(row: PaperRow): DropdownOption[] {
  const draft = row.status === 0
  const published = row.status === 1
  return [
    { label: '编辑信息', key: 'edit', disabled: !draft },
    { label: '手动选题', key: 'pick', disabled: !draft },
    { label: '随机组卷', key: 'random', disabled: !draft },
    { label: '发布', key: 'publish', disabled: !draft || row.question_cnt < 1 },
    { label: '撤销发布', key: 'unpublish', disabled: !published },
    { label: '复制为草稿', key: 'copy', disabled: false },
    { label: '删除草稿', key: 'delete', disabled: !draft },
  ]
}

function onPaperDropdown(key: string, row: PaperRow) {
  switch (key) {
    case 'edit':
      void openEditPaper(row)
      break
    case 'pick':
      openPickModal(row)
      break
    case 'random':
      openRandomModal(row)
      break
    case 'publish':
      dialog.warning({
        title: '发布试卷',
        content: '发布后学生可在开放时间内参加考试。题目总分须与试卷总分一致。',
        positiveText: '发布',
        negativeText: '取消',
        onPositiveClick: async () => {
          try {
            await teacherPaperPublish(row.id)
            message.success('已发布')
            await load()
          } catch (e) {
            message.error(e instanceof Error ? e.message : '发布失败')
          }
        },
      })
      break
    case 'unpublish':
      dialog.warning({
        title: '撤销发布',
        content: '试卷将变回草稿，学生无法再开考。',
        positiveText: '撤销',
        negativeText: '取消',
        onPositiveClick: async () => {
          try {
            await teacherPaperUnpublish(row.id)
            message.success('已撤销发布')
            await load()
          } catch (e) {
            message.error(e instanceof Error ? e.message : '操作失败')
          }
        },
      })
      break
    case 'copy':
      dialog.info({
        title: '复制试卷',
        content: '将复制题目与配置为新草稿（标题加「副本」）。',
        positiveText: '复制',
        negativeText: '取消',
        onPositiveClick: async () => {
          try {
            await teacherPaperCopy(row.id)
            message.success('已复制')
            await load()
          } catch (e) {
            message.error(e instanceof Error ? e.message : '复制失败')
          }
        },
      })
      break
    case 'delete':
      dialog.warning({
        title: '删除试卷',
        content: '仅草稿可删除，且不可恢复。',
        positiveText: '删除',
        negativeText: '取消',
        onPositiveClick: async () => {
          try {
            await teacherPaperDelete(row.id)
            message.success('已删除')
            await load()
          } catch (e) {
            message.error(e instanceof Error ? e.message : '删除失败')
          }
        },
      })
      break
  }
}

const columns: DataTableColumns<PaperRow> = [
  { title: '标题', key: 'title', ellipsis: { tooltip: true } },
  {
    title: '状态',
    key: 'status',
    width: 92,
    render(row) {
      const s = statusMap[row.status] || { label: String(row.status), type: 'default' }
      return h(NTag, { type: s.type, size: 'small' }, { default: () => s.label })
    },
  },
  { title: '题目数', key: 'question_cnt', width: 80 },
  { title: '总分', key: 'total_score', width: 68 },
  {
    title: '考试时间',
    key: 'range',
    width: 200,
    ellipsis: { tooltip: true },
    render(row) {
      return `${row.start_time} ~ ${row.end_time}`
    },
  },
  {
    title: '操作',
    key: 'act',
    width: 280,
    render(row) {
      return h(
        NSpace,
        { size: 'small', wrap: false },
        {
          default: () => [
            h(NButton, { size: 'tiny', onClick: () => openPreview(row) }, { default: () => '预览' }),
            h(
              NButton,
              {
                size: 'tiny',
                type: 'primary',
                secondary: true,
                onClick: () =>
                  router.push({ name: 'teacher-statistics', params: { paperId: String(row.id) } }),
              },
              { default: () => '统计' }
            ),
            h(
              NDropdown,
              {
                trigger: 'click',
                options: paperDropdownOptions(row),
                onSelect: (key: string) => onPaperDropdown(key, row),
              },
              {
                default: () =>
                  h(NButton, { size: 'tiny', secondary: true }, { default: () => '更多 ▾' }),
              }
            ),
          ],
        }
      )
    },
  },
]

async function load() {
  loading.value = true
  try {
    const res = await teacherPaperList(
      page.value,
      pageSize.value,
      keyword.value || undefined,
      status.value
    )
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
  <n-card title="试卷管理">
    <n-space style="margin-bottom: 16px" align="center" wrap>
      <n-input v-model:value="keyword" placeholder="搜索标题" clearable style="width: 200px" @keyup.enter="load" />
      <n-select
        v-model:value="status"
        :options="statusOpts"
        clearable
        placeholder="状态"
        style="width: 140px"
      />
      <n-button type="primary" @click="((page = 1), load())">查询</n-button>
      <n-button type="primary" secondary @click="openCreatePaper">新建试卷</n-button>
    </n-space>
    <n-data-table :columns="columns" :data="items" :loading="loading" />
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

  <n-modal v-model:show="showPaperModal" preset="card" :title="editingPaperId ? '编辑试卷' : '新建试卷'" style="width: 520px">
    <n-form label-placement="top">
      <n-form-item label="标题" required>
        <n-input v-model:value="paperForm.title" placeholder="试卷标题" />
      </n-form-item>
      <n-form-item label="说明">
        <n-input v-model:value="paperForm.description" type="textarea" :rows="2" placeholder="可选" />
      </n-form-item>
      <n-space>
        <n-form-item label="考试时长（分钟）">
          <n-input-number v-model:value="paperForm.duration" :min="1" style="width: 160px" />
        </n-form-item>
        <n-form-item label="试卷总分">
          <n-input-number v-model:value="paperForm.total_score" :min="1" style="width: 160px" />
        </n-form-item>
      </n-space>
      <n-form-item label="开考时间">
        <n-date-picker v-model:value="paperForm.startTs" type="datetime" clearable style="width: 100%" />
      </n-form-item>
      <n-form-item label="结束时间">
        <n-date-picker v-model:value="paperForm.endTs" type="datetime" clearable style="width: 100%" />
      </n-form-item>
    </n-form>
    <n-space justify="end" style="margin-top: 12px">
      <n-button @click="showPaperModal = false">取消</n-button>
      <n-button type="primary" :loading="paperSaving" @click="submitPaperForm">保存</n-button>
    </n-space>
  </n-modal>

  <n-modal
    v-model:show="showPickModal"
    preset="card"
    title="从题库选题加入试卷"
    style="width: min(960px, 98vw)"
    :mask-closable="false"
  >
    <n-space style="margin-bottom: 12px">
      <n-input v-model:value="pickKeyword" placeholder="关键词筛选题库" clearable style="width: 200px" @keyup.enter="((pickPage = 1), loadPickBank())" />
      <n-button size="small" @click="((pickPage = 1), loadPickBank())">查询</n-button>
    </n-space>
    <n-spin :show="pickLoading">
      <n-data-table
        v-model:checked-row-keys="pickChecked"
        :columns="pickColumns"
        :data="pickItems"
        :row-key="(r) => r.id"
      />
    </n-spin>
    <div style="margin-top: 12px; display: flex; justify-content: flex-end">
      <n-pagination
        v-model:page="pickPage"
        v-model:page-size="pickPageSize"
        :item-count="pickTotal"
        show-size-picker
        :page-sizes="[10, 20, 50]"
        @update:page="loadPickBank"
        @update:page-size="
          () => {
            pickPage = 1
            void loadPickBank()
          }
        "
      />
    </div>
    <n-space justify="end" style="margin-top: 16px">
      <n-button @click="showPickModal = false">取消</n-button>
      <n-button type="primary" @click="submitPickQuestions">加入试卷</n-button>
    </n-space>
  </n-modal>

  <n-modal v-model:show="showRandomModal" preset="card" title="随机组卷" style="width: 560px">
    <p class="hint">
      从您的题库中按题型随机抽取。每条规则的合计分将平均到各题（余数分摊到前几题）。成功后试卷「总分」会更新为所有规则合计分；发布前请确认与预期一致。
    </p>
    <div v-for="(rule, idx) in randomRules" :key="idx" class="rule-row">
      <n-select v-model:value="rule.type" :options="typeOptsPick" style="width: 120px" />
      <span>抽</span>
      <n-input-number v-model:value="rule.count" :min="1" style="width: 100px" />
      <span>题，合计</span>
      <n-input-number v-model:value="rule.total_score" :min="1" style="width: 100px" />
      <span>分</span>
      <n-button quaternary size="tiny" @click="removeRandomRule(idx)">删</n-button>
    </div>
    <n-button dashed block style="margin-top: 8px" @click="addRandomRule">添加规则</n-button>
    <n-space justify="end" style="margin-top: 16px">
      <n-button @click="showRandomModal = false">取消</n-button>
      <n-button type="primary" @click="submitRandom">执行随机组卷</n-button>
    </n-space>
  </n-modal>

  <n-modal v-model:show="showPreview" preset="card" :title="previewTitle || '试卷预览'" style="width: min(920px, 96vw)">
    <n-spin :show="previewLoading">
      <div v-for="q in previewQs" :key="q.sort" class="pq">
        <p class="pq-title">{{ q.sort }}. （{{ q.score }} 分）</p>
        <p class="pq-body">{{ q.content }}</p>
        <ul v-if="q.options && parseOptionList(q.options).length" class="pq-opt">
          <li v-for="(o, i) in parseOptionList(q.options)" :key="i">{{ o }}</li>
        </ul>
      </div>
    </n-spin>
  </n-modal>
</template>

<style scoped>
.hint {
  font-size: 13px;
  color: #64748b;
  margin: 0 0 12px;
  line-height: 1.5;
}
.rule-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
  flex-wrap: wrap;
}
.pq {
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
}
.pq-title {
  font-weight: 600;
  margin: 0 0 8px;
}
.pq-body {
  margin: 0 0 8px;
  white-space: pre-wrap;
  line-height: 1.6;
}
.pq-opt {
  margin: 0;
  padding-left: 20px;
}
</style>
