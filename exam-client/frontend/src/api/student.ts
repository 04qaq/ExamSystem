import { http } from './http'
import type { AvailablePaper, PaginatedRecords, StartExamResult } from './types'

export async function studentPapers(): Promise<AvailablePaper[]> {
  const res = await http.get<AvailablePaper[]>('/api/student/papers')
  return res.data
}

export async function studentRecords(page = 1, pageSize = 20): Promise<PaginatedRecords> {
  const res = await http.get<PaginatedRecords>('/api/student/exam/records', {
    params: { page, page_size: pageSize },
  })
  return res.data
}

export async function startExam(paperId: number): Promise<StartExamResult> {
  const res = await http.post<StartExamResult>(`/api/student/papers/${paperId}/start`)
  return res.data
}

export async function saveAnswer(recordId: number, questionId: number, answer: string) {
  await http.post(`/api/student/exam/${recordId}/answer`, {
    question_id: questionId,
    answer,
  })
}

export async function submitExam(recordId: number, tabSwitchCount: number): Promise<SubmitExamStats> {
  const res = await http.post<SubmitExamStats>(`/api/student/exam/${recordId}/submit`, {
    tab_switch_count: tabSwitchCount,
  })
  return res.data
}

export async function recordDetail(recordId: number): Promise<ExamRecordDetail> {
  const res = await http.get<ExamRecordDetail>(`/api/student/exam/records/${recordId}`)
  return res.data
}

export interface AnswerDetailRow {
  question_id: number
  content: string
  type: number
  options: string
  score: number
  provided_answer: string
  correct_answer: string
  is_correct: number | null
  score_gained: number
  comment: string
}

export interface ExamRecordDetail {
  id: number
  paper_title: string
  paper_id: number
  total_score: number
  status: number
  start_time: string
  submit_time: string | null
  details: AnswerDetailRow[]
}

export interface SubmitExamStats {
  total_score: number
  correct_count: number
  wrong_count: number
  submitted_at: string
}
