package controller

import (
	"net/http"
	"strconv"

	"exam-server/internal/dto"
	"exam-server/internal/service"
	"exam-server/pkg/errcode"

	"github.com/gin-gonic/gin"
)

type QuestionController struct {
	questionService *service.QuestionService
}

func NewQuestionController() *QuestionController {
	return &QuestionController{
		questionService: service.NewQuestionService(),
	}
}

func (ctrl *QuestionController) Create(c *gin.Context) {
	var req dto.CreateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, err.Error()))
		return
	}

	teacherID := c.GetUint64("userID")
	resp, code, err := ctrl.questionService.Create(req, teacherID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(resp))
}

func (ctrl *QuestionController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, "无效的题目ID"))
		return
	}

	var req dto.UpdateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, err.Error()))
		return
	}

	teacherID := c.GetUint64("userID")
	resp, code, err := ctrl.questionService.Update(id, req, teacherID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(resp))
}

func (ctrl *QuestionController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, "无效的题目ID"))
		return
	}

	teacherID := c.GetUint64("userID")
	code, err := ctrl.questionService.Delete(id, teacherID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(nil))
}

func (ctrl *QuestionController) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, "无效的题目ID"))
		return
	}

	resp, code, err := ctrl.questionService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(resp))
}

func (ctrl *QuestionController) List(c *gin.Context) {
	var query dto.QuestionQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, err.Error()))
		return
	}

	resp, code, err := ctrl.questionService.List(query)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(resp))
}

func (ctrl *QuestionController) Import(c *gin.Context) {
	var req dto.ImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, err.Error()))
		return
	}

	teacherID := c.GetUint64("userID")
	resp, code, err := ctrl.questionService.Import(req, teacherID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(resp))
}
