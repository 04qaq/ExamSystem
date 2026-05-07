package controller

import (
	"net/http"
	"strconv"

	"exam-server/internal/dto"
	"exam-server/internal/service"
	"exam-server/pkg/errcode"

	"github.com/gin-gonic/gin"
)

type ExamController struct {
	examService *service.ExamService
}

func NewExamController() *ExamController {
	return &ExamController{
		examService: service.NewExamService(),
	}
}

// GetAvailablePapers 可考试卷列表
func (ctrl *ExamController) GetAvailablePapers(c *gin.Context) {
	userID := c.GetUint64("userID")
	resp, code, err := ctrl.examService.GetAvailablePapers(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.Success(resp))
}

// StartExam 开始考试
func (ctrl *ExamController) StartExam(c *gin.Context) {
	paperID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, "无效的试卷ID"))
		return
	}

	userID := c.GetUint64("userID")
	resp, code, err := ctrl.examService.StartExam(paperID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.Success(resp))
}

// SaveAnswer 保存答案
func (ctrl *ExamController) SaveAnswer(c *gin.Context) {
	recordID, err := strconv.ParseUint(c.Param("record_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, "无效的记录ID"))
		return
	}

	var req dto.SaveAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, err.Error()))
		return
	}

	userID := c.GetUint64("userID")
	code, err := ctrl.examService.SaveAnswer(recordID, userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.Success(nil))
}

// SubmitExam 提交试卷
func (ctrl *ExamController) SubmitExam(c *gin.Context) {
	recordID, err := strconv.ParseUint(c.Param("record_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, "无效的记录ID"))
		return
	}

	var req dto.SubmitExamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, err.Error()))
		return
	}

	userID := c.GetUint64("userID")
	resp, code, err := ctrl.examService.SubmitExam(recordID, userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.Success(resp))
}

// GetRecords 考试记录列表
func (ctrl *ExamController) GetRecords(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	userID := c.GetUint64("userID")
	resp, code, err := ctrl.examService.GetRecords(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.Success(resp))
}

// GetRecordDetail 成绩详情
func (ctrl *ExamController) GetRecordDetail(c *gin.Context) {
	recordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, "无效的记录ID"))
		return
	}

	userID := c.GetUint64("userID")
	resp, code, err := ctrl.examService.GetRecordDetail(recordID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.Success(resp))
}
