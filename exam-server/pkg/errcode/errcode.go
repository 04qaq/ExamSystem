package errcode

const (
	Success       = 0
	UnknownError  = 1000
	InvalidParams = 1001

	UserAlreadyExists = 2001
	UserNotFound      = 2002
	PasswordWrong     = 2003
	PasswordTooShort  = 2004

	TokenInvalid   = 3001
	TokenExpired   = 3002
	NoPermission   = 4003

	QuestionNotFound      = 5001
	ImportFailed          = 5002
	PaperNotFound         = 6001
	PaperNotDraft         = 6002
	PaperNoQuestions      = 6003
	PaperScoreMismatch    = 6004
	PaperCannotModify     = 6005
	InsufficientQuestions = 6006
	PaperNotPublished     = 6007
	PaperTimeInvalid      = 6008
	ExamRecordNotFound    = 7001
	ExamAlreadySubmitted  = 7002
	ExamTimeExpired       = 7003
	ExamInProgress        = 7004
	ExamNotSubmitted      = 7005
	GradeAlreadyDone      = 7006
	UserDisabled          = 8001
	UserCreateFailed      = 8002
	UserUpdateFailed      = 8003
	LogNotFound           = 8004
)

var messages = map[int]string{
	Success:       "成功",
	UnknownError:  "未知错误",
	InvalidParams: "参数无效",

	UserAlreadyExists: "用户名已存在",
	UserNotFound:      "用户不存在",
	PasswordWrong:     "密码错误",
	PasswordTooShort:  "密码长度至少6位",

	TokenInvalid:   "Token无效",
	TokenExpired:   "Token已过期",
	NoPermission:   "无权限",

	QuestionNotFound: "题目不存在",
	ImportFailed:     "导入失败",
	PaperNotFound:    "试卷不存在",
	PaperNotDraft:    "仅草稿状态的试卷可操作",
	PaperNoQuestions: "试卷中无题目",
	PaperScoreMismatch: "题目总分与试卷总分不一致",
	PaperCannotModify:  "试卷当前状态不允许修改",
	InsufficientQuestions: "符合条件的题目数量不足",
	PaperNotPublished:     "试卷未发布",
	PaperTimeInvalid:      "不在考试有效时间内",
	ExamRecordNotFound:    "考试记录不存在",
	ExamAlreadySubmitted:  "试卷已提交",
	ExamTimeExpired:       "考试时间已结束",
	ExamInProgress:        "该试卷已有进行中的考试记录",
	ExamNotSubmitted:      "考试尚未提交",
	GradeAlreadyDone:      "该题目已批阅",
	UserDisabled:          "用户已被禁用",
	UserCreateFailed:      "创建用户失败",
	UserUpdateFailed:      "更新用户失败",
	LogNotFound:           "日志不存在",
}

func Message(code int) string {
	if msg, ok := messages[code]; ok {
		return msg
	}
	return "未知错误"
}
