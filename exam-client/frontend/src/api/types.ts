/** 与 exam-server 约定一致 */

export interface ApiEnvelope<T = unknown> {
  code: number
  message: string
  data?: T
}

export const RoleAdmin = 1
export const RoleTeacher = 2
export const RoleStudent = 3

export interface UserInfo {
  id: number
  username: string
  role: number
  real_name: string
}

export interface LoginBody {
  username: string
  password: string
}

export interface LoginResult {
  token: string
  refresh_token: string
  expires_in: number
  user: UserInfo
}

export interface RegisterBody {
  username: string
  password: string
  real_name?: string
}

export interface AvailablePaper {
  id: number
  title: string
  description: string
  duration: number
  total_score: number
  start_time: string
  end_time: string
  my_record_id?: number
}

export interface ExamQuestionItem {
  sort_order: number
  question_id: number
  type: number
  content: string
  options: string
  score: number
}

export interface StartExamResult {
  record_id: number
  questions: ExamQuestionItem[]
  remaining_seconds: number
  total_score: number
}

export interface ExamRecordRow {
  id: number
  paper_title: string
  paper_id: number
  total_score: number
  status: number
  start_time: string
  submit_time: string | null
}

export interface PaginatedRecords {
  total: number
  items: ExamRecordRow[]
}

export type QuestionType = 1 | 2 | 3 | 4 | 5

export interface PaperRow {
  id: number
  title: string
  description: string
  duration: number
  total_score: number
  start_time: string
  end_time: string
  status: number
  creator_id: number
  question_cnt: number
  created_at: string
  updated_at: string
}

export interface PaginatedPapers {
  total: number
  items: PaperRow[]
}

export interface QuestionRow {
  id: number
  type: number
  content: string
  options: string
  answer: string
  score: number
  difficulty: number
  tags: string
  creator_id: number
  created_at: string
  updated_at: string
}

export interface QuestionListResponse {
  total: number
  items: QuestionRow[]
}

export interface UserRow {
  id: number
  username: string
  role: number
  real_name: string
  status: number
  created_at: string
}
