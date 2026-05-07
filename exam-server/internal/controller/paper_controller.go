package controller

import (
	"net/http"
	"strconv"

	"exam-server/internal/dto"
	"exam-server/internal/service"
	"exam-server/pkg/errcode"

	"github.com/gin-gonic/gin"
)

type PaperController struct {
	paperService *service.PaperService
}

func NewPaperController() *PaperController {
	return &PaperController{
		paperService: service.NewPaperService(),
	}
}

func (ctrl *PaperController) Create(c *gin.Context) {
	var req dto.CreatePaperRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, err.Error()))
		return
	}

	teacherID := c.GetUint64("userID")
	resp, code, err := ctrl.paperService.Create(req, teacherID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(resp))
}

func (ctrl *PaperController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, "无效的试卷ID"))
		return
	}

	var req dto.UpdatePaperRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, err.Error()))
		return
	}

	teacherID := c.GetUint64("userID")
	resp, code, err := ctrl.paperService.Update(id, req, teacherID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(resp))
}

func (ctrl *PaperController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, "无效的试卷ID"))
		return
	}

	teacherID := c.GetUint64("userID")
	code, err := ctrl.paperService.Delete(id, teacherID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(nil))
}

func (ctrl *PaperController) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, "无效的试卷ID"))
		return
	}

	teacherID := c.GetUint64("userID")
	resp, code, err := ctrl.paperService.GetByID(id, teacherID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(resp))
}

func (ctrl *PaperController) List(c *gin.Context) {
	var query dto.PaperQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, err.Error()))
		return
	}

	teacherID := c.GetUint64("userID")
	resp, code, err := ctrl.paperService.List(query, teacherID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(resp))
}

func (ctrl *PaperController) AddQuestions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, "无效的试卷ID"))
		return
	}

	var req dto.AddQuestionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, err.Error()))
		return
	}

	teacherID := c.GetUint64("userID")
	resp, code, err := ctrl.paperService.AddQuestions(id, req, teacherID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(resp))
}

func (ctrl *PaperController) RandomSelect(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, "无效的试卷ID"))
		return
	}

	var req dto.RandomSelectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, err.Error()))
		return
	}

	teacherID := c.GetUint64("userID")
	resp, code, err := ctrl.paperService.RandomSelect(id, req, teacherID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(resp))
}

func (ctrl *PaperController) Publish(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, "无效的试卷ID"))
		return
	}

	teacherID := c.GetUint64("userID")
	resp, code, err := ctrl.paperService.Publish(id, teacherID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(resp))
}

func (ctrl *PaperController) Unpublish(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, "无效的试卷ID"))
		return
	}

	teacherID := c.GetUint64("userID")
	resp, code, err := ctrl.paperService.Unpublish(id, teacherID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(resp))
}

func (ctrl *PaperController) Copy(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, "无效的试卷ID"))
		return
	}

	teacherID := c.GetUint64("userID")
	resp, code, err := ctrl.paperService.Copy(id, teacherID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(resp))
}

func (ctrl *PaperController) Preview(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, "无效的试卷ID"))
		return
	}

	teacherID := c.GetUint64("userID")
	resp, code, err := ctrl.paperService.Preview(id, teacherID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(resp))
}
