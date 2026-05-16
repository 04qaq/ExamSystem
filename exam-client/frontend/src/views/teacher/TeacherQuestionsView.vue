<script setup lang="ts">
import {
  NCard,
  NDataTable,
  NPagination,
  NSpace,
  NButton,
  NSelect,
  NInput,
  NModal,
  NForm,
  NFormItem,
  NInputNumber,
  useMessage,
  useDialog,
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { h, onMounted, reactive, ref } from 'vue'
import type { QuestionRow } from '@/api/types'
import {
  teacherQuestionCreate,
  teacherQuestionDelete,
  teacherQuestionGet,
  teacherQuestionImport,
  teacherQuestionList,
  teacherQuestionUpdate,
  type QuestionPayload,
} from '@/api/teacher'
import { normalizeQuestionOptionsInput } from '@/utils/options'

const message = useMessage()
const dialog = useDialog()

const loading = ref(false)
const items = ref<QuestionRow[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const qType = ref<number | undefined>(undefined)
const difficulty = ref<number | undefined>(undefined)
const keyword = ref('')
const tag = ref('')

const typeOpts = [
  { label: '单选', value: 1 },
  { label: '多选', value: 2 },
  { label: '判断', value: 3 },
  { label: '填空', value: 4 },
  { label: '简答', value: 5 },
]

const diffOpts = [
  { label: '易', value: 1 },
  { label: '中', value: 2 },
  { label: '难', value: 3 },
]

const typeLabel: Record<number, string> = { 1: '单选', 2: '多选', 3: '判断', 4: '填空', 5: '简答' }

/** 接口返回的 options 多为「换行分隔」；旧数据可能仍是 JSON 数组字符串，编辑时展开为多行 */
function optionsApiToEditorLines(raw: string): string {
  if (!raw?.trim()) return ''
  try {
    const j = JSON.parse(raw) as unknown
    if (Array.isArray(j)) return j.map(String).join('\n')
  } catch {
    /* fallthrough */
  }
  return raw.replace(/\r\n/g, '\n')
}

function emptyForm(): QuestionPayload & { optionsLines: string } {
  return {
    type: 1,
    content: '',
    options: '',
    optionsLines: '',
    answer: '',
    score: 5,
    difficulty: 1,
    tags: '',
  }
}

const showEditor = ref(false)
const showImport = ref(false)
const saving = ref(false)
const editingId = ref<number | null>(null)
const formModel = reactive(emptyForm())

const importText = ref(`[
  {
    "type": 1,
    "content": "2+2=?",
    "options": "4\\n5",
    "answer": "4",
    "score": 5,
    "difficulty": 1,
    "tags": ""
  }
]`)

async function load() {
  loading.value = true
  try {
    const res = await teacherQuestionList(page.value, pageSize.value, {
      type: qType.value,
      difficulty: difficulty.value,
      keyword: keyword.value || undefined,
      tag: tag.value || undefined,
    })
    items.value = res.items
    total.value = res.total
  } catch (e) {
    message.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = null
  Object.assign(formModel, emptyForm())
  showEditor.value = true
}

async function openEdit(row: QuestionRow) {
  editingId.value = row.id
  try {
    const q = await teacherQuestionGet(row.id)
    formModel.type = q.type
    formModel.content = q.content
    formModel.answer = q.answer
    formModel.score = q.score
    formModel.difficulty = q.difficulty
    formModel.tags = q.tags || ''
    formModel.options = q.options || ''
    formModel.optionsLines = optionsApiToEditorLines(q.options || '')
    showEditor.value = true
  } catch (e) {
    message.error(e instanceof Error ? e.message : '加载题目失败')
  }
}

function payloadFromForm(): QuestionPayload {
  let opts = ''
  if (formModel.type === 1 || formModel.type === 2) {
    opts = normalizeQuestionOptionsInput(formModel.optionsLines)
  } else if (formModel.type === 3) {
    opts = normalizeQuestionOptionsInput(formModel.optionsLines)
    if (!opts.trim()) opts = '正确\n错误'
  } else {
    opts = normalizeQuestionOptionsInput(formModel.options)
  }
  return {
    type: formModel.type,
    content: formModel.content.trim(),
    options: opts,
    answer: formModel.answer.trim(),
    score: formModel.score,
    difficulty: formModel.difficulty,
    tags: formModel.tags.trim(),
  }
}

async function submitForm() {
  const p = payloadFromForm()
  if ((p.type === 1 || p.type === 2) && !p.options.trim()) {
    message.warning('请填写选项（每行一项；勿整段 JSON.stringify）')
    return
  }
  saving.value = true
  try {
    if (editingId.value != null) {
      await teacherQuestionUpdate(editingId.value, p)
      message.success('已保存')
    } else {
      await teacherQuestionCreate(p)
      message.success('已创建')
    }
    showEditor.value = false
    await load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '保存失败')
  } finally {
    saving.value = false
  }
}

function confirmDelete(row: QuestionRow) {
  dialog.warning({
    title: '删除题目',
    content: `确定删除题目 #${row.id} 吗？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await teacherQuestionDelete(row.id)
        message.success('已删除')
        await load()
      } catch (e) {
        message.error(e instanceof Error ? e.message : '删除失败')
      }
    },
  })
}

async function submitImport() {
  let parsed: unknown
  try {
    parsed = JSON.parse(importText.value.trim())
  } catch {
    message.error('JSON 格式不正确')
    return
  }
  const arr = Array.isArray(parsed) ? parsed : (parsed as { questions?: unknown }).questions
  if (!Array.isArray(arr) || arr.length === 0) {
    message.error('请提供题目数组，或用 { "questions": [...] } 包裹')
    return
  }
  const questions: QuestionPayload[] = []
  for (const raw of arr) {
    if (!raw || typeof raw !== 'object') continue
    const o = raw as Record<string, unknown>
    const t = Number(o.type)
    let optStr = String(o.options ?? '')
    if (t === 1 || t === 2) {
      optStr = normalizeQuestionOptionsInput(optStr)
    } else if (t === 3) {
      const n = normalizeQuestionOptionsInput(optStr)
      optStr = n.trim() ? n : '正确\n错误'
    } else {
      optStr = normalizeQuestionOptionsInput(optStr)
    }
    questions.push({
      type: t,
      content: String(o.content ?? ''),
      options: optStr,
      answer: String(o.answer ?? ''),
      score: Number(o.score) || 5,
      difficulty: Number(o.difficulty) || 1,
      tags: String(o.tags ?? ''),
    })
  }
  if (!questions.length) {
    message.error('没有有效题目')
    return
  }
  saving.value = true
  try {
    const res = await teacherQuestionImport(questions)
    message.success(`导入完成：成功 ${res.success}，失败 ${res.fail}，共 ${res.total}`)
    showImport.value = false
    await load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '导入失败')
  } finally {
    saving.value = false
  }
}

const columns: DataTableColumns<QuestionRow> = [
  { title: 'ID', key: 'id', width: 70 },
  {
    title: '题型',
    key: 'type',
    width: 72,
    render(row) {
      return typeLabel[row.type] || row.type
    },
  },
  { title: '分值', key: 'score', width: 64 },
  {
    title: '难度',
    key: 'difficulty',
    width: 52,
    render(row) {
      return ({ 1: '易', 2: '中', 3: '难' } as Record<number, string>)[row.difficulty] || row.difficulty
    },
  },
  {
    title: '题干',
    key: 'content',
    ellipsis: { tooltip: true },
  },
  {
    title: '操作',
    key: 'act',
    width: 140,
    render(row) {
      return h(NSpace, null, {
        default: () => [
          h(NButton, { size: 'tiny', onClick: () => openEdit(row) }, { default: () => '编辑' }),
          h(
            NButton,
            { size: 'tiny', type: 'error', secondary: true, onClick: () => confirmDelete(row) },
            { default: () => '删除' }
          ),
        ],
      })
    },
  },
]

onMounted(load)
</script>

<template>
  <n-card title="题库">
    <n-space style="margin-bottom: 16px" align="center" wrap>
      <n-select v-model:value="qType" :options="typeOpts" clearable placeholder="题型" style="width: 130px" />
      <n-select v-model:value="difficulty" :options="diffOpts" clearable placeholder="难度" style="width: 130px" />
      <n-input v-model:value="keyword" placeholder="关键词" clearable style="width: 160px" @keyup.enter="load" />
      <n-input v-model:value="tag" placeholder="标签" clearable style="width: 120px" @keyup.enter="load" />
      <n-button type="primary" @click="((page = 1), load())">查询</n-button>
      <n-button @click="openCreate">新建题目</n-button>
      <n-button secondary @click="showImport = true">批量导入 JSON</n-button>
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

  <n-modal
    v-model:show="showEditor"
    preset="card"
    :title="editingId ? `编辑题目 #${editingId}` : '新建题目'"
    style="width: min(640px, 96vw)"
    :mask-closable="false"
  >
    <n-form :model="formModel" label-placement="top">
      <n-form-item label="题型">
        <n-select v-model:value="formModel.type" :options="typeOpts" />
      </n-form-item>
      <n-form-item label="题干">
        <n-input v-model:value="formModel.content" type="textarea" :rows="4" placeholder="题目内容" />
      </n-form-item>
      <n-form-item
        v-if="formModel.type === 1 || formModel.type === 2 || formModel.type === 3"
        label="选项"
      >
        <n-input v-model:value="formModel.optionsLines" type="textarea" :rows="5" placeholder="每行一个选项，例如：&#10;A. 选项一&#10;B. 选项二" />
        <p class="field-tip">单选、多选、判断题会按行识别选项；判断题不填时默认使用“正确 / 错误”。</p>
      </n-form-item>
      <n-form-item v-else label="选项">
        <n-input v-model:value="formModel.options" type="textarea" :rows="2" placeholder="可选" />
        <p class="field-tip">填空题和简答题通常不用填写选项。</p>
      </n-form-item>
      <n-form-item label="答案">
        <n-input v-model:value="formModel.answer" placeholder="单选/判断：与选项文案一致；多选：逗号分隔" />
        <p class="field-tip">客观题答案需要与选项文字一致；多选题用英文逗号分隔，如 A,C。</p>
      </n-form-item>
      <n-form-item label="分值">
        <n-input-number v-model:value="formModel.score" :min="1" style="width: 160px" />
      </n-form-item>
      <n-form-item label="难度">
        <n-select v-model:value="formModel.difficulty" :options="diffOpts" style="width: 160px" />
      </n-form-item>
      <n-form-item label="标签">
        <n-input v-model:value="formModel.tags" placeholder="可选" />
      </n-form-item>
    </n-form>
    <n-space justify="end" style="margin-top: 12px">
      <n-button @click="showEditor = false">取消</n-button>
      <n-button type="primary" :loading="saving" @click="submitForm">保存</n-button>
    </n-space>
  </n-modal>

  <n-modal v-model:show="showImport" preset="card" title="批量导入题目（JSON）" style="width: min(720px, 96vw)">
    <p class="hint">
      提交体为题目对象数组，字段与创建接口一致：<code>type, content, options, answer, score, difficulty, tags</code>。
      单选/多选/判断的 <code>options</code> 请用换行分隔选项（字符串内写 <code>\n</code>），也可直接粘贴 JSON 数组，导入时会转为多行再提交。
    </p>
    <n-input v-model:value="importText" type="textarea" :rows="14" class="mono" />
    <n-space justify="end" style="margin-top: 12px">
      <n-button @click="showImport = false">取消</n-button>
      <n-button type="primary" :loading="saving" @click="submitImport">导入</n-button>
    </n-space>
  </n-modal>
</template>

<style scoped>
.hint {
  font-size: 13px;
  color: #64748b;
  margin: 0 0 12px;
  line-height: 1.5;
}
.field-tip {
  margin: 6px 0 0;
  font-size: 12px;
  color: #64748b;
  line-height: 1.5;
}
.mono :deep(textarea) {
  font-family: ui-monospace, monospace;
  font-size: 12px;
}
code {
  padding: 1px 6px;
  border-radius: 4px;
  background: rgba(0, 0, 0, 0.06);
}
</style>
