package router

import (
	"exam-server/internal/controller"
	"exam-server/internal/middleware"
	"exam-server/internal/model"

	"github.com/gin-gonic/gin"
)

func Setup() *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS())

	authCtrl := controller.NewAuthController()
	questionCtrl := controller.NewQuestionController()
	paperCtrl := controller.NewPaperController()
	examCtrl := controller.NewExamController()
	gradeCtrl := controller.NewGradeController()
	adminCtrl := controller.NewAdminController()

	// 公开路由
	public := r.Group("/api")
	{
		public.POST("/auth/register", authCtrl.Register)
		public.POST("/auth/login", authCtrl.Login)
	}

	// 需要认证的路由
	auth := r.Group("/api")
	auth.Use(middleware.JWTAuth())
	{
		// 管理员路由
		admin := auth.Group("/admin")
		admin.Use(middleware.RequireRole(model.RoleAdmin))
		{
			admin.GET("/users", adminCtrl.ListUsers)
			admin.POST("/users", adminCtrl.CreateUser)
			admin.PUT("/users/:id", adminCtrl.UpdateUser)
			admin.GET("/logs", adminCtrl.ListLogs)
		}

		// 教师路由
		teacher := auth.Group("/teacher")
		teacher.Use(middleware.RequireRole(model.RoleTeacher))
		{
			teacher.GET("/questions", questionCtrl.List)
			teacher.POST("/questions", questionCtrl.Create)
			teacher.GET("/questions/:id", questionCtrl.GetByID)
			teacher.PUT("/questions/:id", questionCtrl.Update)
			teacher.DELETE("/questions/:id", questionCtrl.Delete)
			teacher.POST("/questions/import", questionCtrl.Import)

			teacher.GET("/papers", paperCtrl.List)
			teacher.POST("/papers", paperCtrl.Create)
			teacher.GET("/papers/:id", paperCtrl.GetByID)
			teacher.PUT("/papers/:id", paperCtrl.Update)
			teacher.DELETE("/papers/:id", paperCtrl.Delete)
			teacher.POST("/papers/:id/questions", paperCtrl.AddQuestions)
			teacher.POST("/papers/:id/random-select", paperCtrl.RandomSelect)
			teacher.PUT("/papers/:id/publish", paperCtrl.Publish)
			teacher.PUT("/papers/:id/unpublish", paperCtrl.Unpublish)
			teacher.POST("/papers/:id/copy", paperCtrl.Copy)
			teacher.GET("/papers/:id/preview", paperCtrl.Preview)

			// 成绩批阅与统计
			teacher.GET("/exam/pending", gradeCtrl.GetPendingList)
			teacher.GET("/exam/records/:id/grade", gradeCtrl.GetGradeDetail)
			teacher.POST("/exam/records/:id/grade", gradeCtrl.GradeSubmit)
			teacher.GET("/statistics/paper/:id", gradeCtrl.GetPaperStatistics)
			teacher.GET("/statistics/paper/:id/export", gradeCtrl.ExportExcel)
		}

		// 学生路由
		student := auth.Group("/student")
		student.Use(middleware.RequireRole(model.RoleStudent))
		{
			student.GET("/papers", examCtrl.GetAvailablePapers)
			student.POST("/papers/:id/start", examCtrl.StartExam)
			student.POST("/exam/:record_id/answer", examCtrl.SaveAnswer)
			student.POST("/exam/:record_id/submit", examCtrl.SubmitExam)
			student.GET("/exam/records", examCtrl.GetRecords)
			student.GET("/exam/records/:id", examCtrl.GetRecordDetail)
		}
	}

	return r
}
