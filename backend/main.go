package main

import (
	"context"
	"breakout.cn/backend/internal/controller"
	"breakout.cn/backend/internal/service"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "Breakout backend server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server()

			// 初始化服务
			service.Init()

			// 全局中间件
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(
					ghttp.MiddlewareCORS,
					service.Middleware.ResponseHandler,
				)

				// 公开接口
				group.POST("/api/v1/user/login", controller.User.Login)
				group.POST("/api/v1/user/register", controller.User.Register)
				group.POST("/api/v1/user/wx-login", controller.User.WxLogin)

				// 需要认证的接口
				group.Group("/api/v1", func(group *ghttp.RouterGroup) {
					group.Middleware(service.Middleware.Auth)

					// 用户相关
					group.GET("/user/info", controller.User.Info)
					group.PUT("/user/profile", controller.User.UpdateProfile)

					// 团队相关
					group.GET("/team", controller.Team.Get)
					group.POST("/team", controller.Team.Post)
					group.PUT("/team", controller.Team.Put)
					group.DELETE("/team", controller.Team.Delete)
					group.POST("/team/invite", controller.Team.InviteMember)
					group.DELETE("/team/member/{id}", controller.Team.RemoveMember)

					// 项目相关
					group.GET("/project", controller.Project.List)
					group.GET("/project/{id}", controller.Project.Get)
					group.POST("/project", controller.Project.Post)
					group.PUT("/project/{id}", controller.Project.Put)
					group.DELETE("/project/{id}", controller.Project.Delete)
					group.GET("/project/{id}/members", controller.Project.GetMembers)
					group.POST("/project/{id}/members", controller.Project.AddMember)
					group.DELETE("/project/{id}/member/{user_id}", controller.Project.RemoveMember)

					// 任务相关（静态路由优先注册）
					group.GET("/task", controller.Task.List)
					group.GET("/task/tree", controller.Task.GetTree)
					group.GET("/task/blocked", controller.Task.GetBlocked)
					group.GET("/task/{id}", controller.Task.Get)
					group.POST("/task", controller.Task.Post)
					group.PUT("/task/{id}", controller.Task.Put)
					group.PUT("/task/{id}/status", controller.Task.UpdateStatus)
					group.DELETE("/task/{id}", controller.Task.Delete)
					group.POST("/task/{id}/comment", controller.Task.AddComment)
					group.POST("/task/{id}/block", controller.Task.MarkBlocked)
					group.POST("/task/{id}/unblock", controller.Task.MarkUnblocked)
					group.POST("/task/{id}/review", controller.Task.Review)
					group.POST("/task/{id}/accept", controller.Task.Accept)
					group.POST("/task/{id}/close", controller.Task.Close)
					group.POST("/task/relation", controller.Task.AddRelation)
					group.DELETE("/task/relation/{id}", controller.Task.RemoveRelation)

					// 评论相关
					group.GET("/comment", controller.Comment.List)
					group.POST("/comment", controller.Comment.Post)
					group.DELETE("/comment/{id}", controller.Comment.Delete)
					group.GET("/comment/{id}/replies", controller.Comment.GetReplies)
					group.POST("/comment/{id}/reply", controller.Comment.CreateReply)

					// 附件相关
					group.POST("/attachment/upload", controller.Attachment.Upload)
					group.GET("/attachment", controller.Attachment.List)
					group.GET("/attachment/stats", controller.Attachment.GetStats)
					group.GET("/attachment/{id}/download", controller.Attachment.Download)
					group.GET("/attachment/{id}/preview", controller.Attachment.Preview)
					group.DELETE("/attachment/{id}", controller.Attachment.Delete)

					// 通知相关
					group.GET("/notification", controller.Notification.List)
					group.GET("/notification/unread-count", controller.Notification.GetUnreadCount)
					group.POST("/notification/read-all", controller.Notification.MarkAllRead)
					group.POST("/notification/clear-all", controller.Notification.ClearAll)
					group.POST("/notification/{id}/read", controller.Notification.MarkRead)
					group.DELETE("/notification/{id}", controller.Notification.Delete)

					// AI相关
					group.POST("/ai/decompose-task", controller.AI.DecomposeTask)
					group.POST("/ai/generate-task-desc", controller.AI.GenerateTaskDescription)
					group.POST("/ai/suggest-assignee", controller.AI.SuggestAssignee)
					group.GET("/ai/usage-stats", controller.AI.GetUsageStats)

					// 报表相关
					group.GET("/report/weekly", controller.Report.GetWeeklyList)
					group.GET("/report/weekly/{id}", controller.Report.GetWeeklyReport)
					group.POST("/report/weekly/generate", controller.Report.GenerateWeeklyReport)
					group.GET("/report/statistics", controller.Report.GetStatistics)
					group.GET("/report/gantt", controller.Report.GetGanttChart)
				})
			})

			s.SetPort(8080)
			s.Run()
			return nil
		},
	}
)

func main() {
	Main.Run(context.Background())
}