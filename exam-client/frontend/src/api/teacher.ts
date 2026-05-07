import { http } from './http'
import type { PaginatedPapers, PaperRow, QuestionListResponse, QuestionRow } from './types'

export interface PaperSavePayload {
  title: string
  description: string
  duration: number
  total_score: number
  start_time: string
  end_time: string
}

export async function teacherPaperGet(id: number): Promise<PaperRow> {
  const res = await http.get<PaperRow>(`/api/teacher/papers/${id}`)
  return res.data
}

export async function teacherPaperCreate(body: PaperSavePayload): Promise<PaperRow> {
  const res = await http.post<PaperRow>('/api/teacher/papers', body)
  return res.data
}

export async function teacherPaperUpdate(id: number, body: PaperSavePayload): Promise<PaperRow> {
  const res = await http.put<PaperRow>(`/api/teacher/papers/${id}`, body)
  return res.data
}

export async function teacherPaperDelete(id: number) {
  await http.delete(`/api/teacher/papers/${id}`)
}

export async function teacherPaperAddQuestions(
  paperId: number,
  questions: { question_id: number; score: number }[]
): Promise<PaperRow> {
  const res = await http.post<PaperRow>(`/api/teacher/papers/${paperId}/questions`, { questions })
  return res.data
}

export async function teacherPaperRandomSelect(
  paperId: number,
  rules: { type: number; count: number; total_score: number }[]
): Promise<PaperRow> {
  const res = await http.post<PaperRow>(`/api/teacher/papers/${paperId}/random-select`, { rules })
  return res.data
}

export async function teacherPaperPublish(id: number): Promise<PaperRow> {
  const res = await http.put<PaperRow>(`/api/teacher/papers/${id}/publish`)
  return res.data
}

export async function teacherPaperUnpublish(id: number): Promise<PaperRow> {
  const res = await http.put<PaperRow>(`/api/teacher/papers/${id}/unpublish`)
  return res.data
}

export async function teacherPaperCopy(id: number): Promise<PaperRow> {
  const res = await http.post<PaperRow>(`/api/teacher/papers/${id}/copy`)
  return res.data
}

export interface QuestionPayload {
  type: number
  content: string
  options: string
  answer: string
  score: number
  difficulty: number
  tags: string
}

export async function teacherQuestionGet(id: number): Promise<QuestionRow> {
  const res = await http.get<QuestionRow>(`/api/teacher/questions/${id}`)
  return res.data
}

export async function teacherQuestionCreate(body: QuestionPayload): Promise<QuestionRow> {
  const res = await http.post<QuestionRow>('/api/teacher/questions', body)
  return res.data
}

export async function teacherQuestionUpdate(id: number, body: QuestionPayload): Promise<QuestionRow> {
  const res = await http.put<QuestionRow>(`/api/teacher/questions/${id}`, body)
  return res.data
}

export async function teacherQuestionDelete(id: number) {
  await http.delete(`/api/teacher/questions/${id}`)
}

export async function teacherQuestionImport(questions: QuestionPayload[]) {
  const res = await http.post<{ total: number; success: number; fail: number }>(
    '/api/teacher/questions/import',
    { questions }
  )
  return res.data
}

export async function teacherPaperList(
  page: number,
  pageSize: number,
  keyword?: string,
  status?: number
): Promise<PaginatedPapers> {
  const res = await http.get<PaginatedPapers>('/api/teacher/papers', {
    params: { page, page_size: pageSize, keyword: keyword || undefined, status },
  })
  return res.data
}

export async function teacherQuestionList(
  page: number,
  pageSize: number,
  filters?: { type?: number; difficulty?: number; tag?: string; keyword?: string }
): Promise<QuestionListResponse> {
  const res = await http.get<QuestionListResponse>('/api/teacher/questions', {
    params: {
      page,
      page_size: pageSize,
      type: filters?.type,
      difficulty: filters?.difficulty,
      tag: filters?.tag,
      keyword: filters?.keyword,
    },
  })
  return res.data
}

export interface PendingRow {
  record_id: number
  paper_id: number
  paper_title: string
  student_name: string
  student_id: number
  submit_time: string
  subjective_count: number
  graded_count: number
  total_score: number
  status: number
}

export interface PendingList {
  total: number
  items: PendingRow[]
}

export async function pendingGrades(page = 1, pageSize = 20): Promise<PendingList> {
  const res = await http.get<PendingList>('/api/teacher/exam/pending', {
    params: { page, page_size: pageSize },
  })
  return res.data
}

export interface GradeDetailItem {
  detail_id: number
  question_id: number
  type: number
  content: string
  options: string
  score: number
  provided_answer: string
  correct_answer: string
  is_correct: number | null
  score_gained: number
  comment: string
}

export interface GradeDetail {
  record_id: number
  paper_title: string
  student_name: string
  total_score: number
  status: number
  submit_time: string
  details: GradeDetailItem[]
}

export async function gradeDetail(recordId: number): Promise<GradeDetail> {
  const res = await http.get<GradeDetail>(`/api/teacher/exam/records/${recordId}/grade`)
  return res.data
}

export async function gradeSubmit(
  recordId: number,
  grades: { detail_id: number; score_gained: number; comment: string }[]
) {
  await http.post(`/api/teacher/exam/records/${recordId}/grade`, { grades })
}

export interface StatisticsPayload {
  paper: {
    paper_id: number
    paper_title: string
    total_students: number
    submitted_count: number
    avg_score: number
    max_score: number
    min_score: number
    pass_count: number
    fail_count: number
    pass_rate: number
  }
  question_stats: {
    question_id: number
    content: string
    type: number
    total_count: number
    correct_count: number
    correct_rate: number
    avg_score: number
    full_score: number
  }[]
  score_distribution: Record<string, number>
}

export async function paperStatistics(paperId: number): Promise<StatisticsPayload> {
  const res = await http.get<StatisticsPayload>(`/api/teacher/statistics/paper/${paperId}`)
  return res.data
}

export interface PaperPreview {
  paper: PaperRow
  questions: {
    id: number
    paper_id: number
    question_id: number
    sort_order: number
    score: number
    question_type?: number
    question_content?: string
    question_options?: string
  }[]
}

export async function paperPreview(paperId: number): Promise<PaperPreview> {
  const res = await http.get<PaperPreview>(`/api/teacher/papers/${paperId}/preview`)
  return res.data
}
