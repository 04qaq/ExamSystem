package controller

import (
	"net/http"
	"strconv"

	"exam-server/internal/dto"
	"exam-server/internal/service"
	"exam-server/pkg/errcode"

	"github.com/gin-gonic/gin"
)

type GradeController struct {
	gradeService *service.GradeService
}

func NewGradeController() *GradeController {
	return &GradeController{
		gradeService: service.NewGradeService(),
	}
}

// GetPendingList 待批阅列表
func (ctrl *GradeController) GetPendingList(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	teacherID := c.GetUint64("userID")
	resp, code, err := ctrl.gradeService.GetPendingList(teacherID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.Success(resp))
}

// GetGradeDetail 批阅详情
func (ctrl *GradeController) GetGradeDetail(c *gin.Context) {
	recordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, "无效的记录ID"))
		return
	}

	teacherID := c.GetUint64("userID")
	resp, code, err := ctrl.gradeService.GetGradeDetail(recordID, teacherID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.Success(resp))
}

// GradeSubmit 提交批阅
func (ctrl *GradeController) GradeSubmit(c *gin.Context) {
	recordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, "无效的记录ID"))
		return
	}

	var req dto.GradeBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, err.Error()))
		return
	}

	teacherID := c.GetUint64("userID")
	code, err := ctrl.gradeService.GradeSubmit(recordID, teacherID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.Success(nil))
}

// GetPaperStatistics 试卷统计
func (ctrl *GradeController) GetPaperStatistics(c *gin.Context) {
	paperID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, "无效的试卷ID"))
		return
	}

	teacherID := c.GetUint64("userID")
	resp, code, err := ctrl.gradeService.GetPaperStatistics(paperID, teacherID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.Success(resp))
}

// ExportExcel 导出成绩 Excel
func (ctrl *GradeController) ExportExcel(c *gin.Context) {
	paperID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, "无效的试卷ID"))
		return
	}

	teacherID := c.GetUint64("userID")
	data, filename, code, err := ctrl.gradeService.ExportExcel(paperID, teacherID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
