package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/trustwired/internal-t/internal/config"
	"github.com/trustwired/internal-t/internal/handlers"
	"github.com/trustwired/internal-t/internal/middleware"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, h *handlers.Handler, db *gorm.DB, cfg *config.Config) {
	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins(),
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length", "Content-Disposition"},
		AllowCredentials: true,
	}))

	authMiddleware := middleware.AuthRequired(db)
	requireAdmin := middleware.RequireRoles("admin")
	requireSupportOrAdmin := middleware.RequireRoles("support", "admin")
	requireSalesOrAdmin := middleware.RequireRoles("sales", "admin")
	requireCS := middleware.RequireRoles("cs", "admin")

	api := r.Group("/api")
	{
		// Public auth routes
		api.POST("/login", h.Login)
		api.POST("/magic-link", h.RequestMagicLink)
		api.POST("/magic-link/verify", h.VerifyMagicLink)
		api.POST("/forgot-password", h.ForgotPassword)
		api.POST("/reset-password", h.ResetPassword)
		api.POST("/email/resend-verification", h.ResendVerification)
		api.GET("/email/verify/:id/:hash", h.VerifyEmail)

		// Protected routes
		protected := api.Group("")
		protected.Use(authMiddleware)
		{
			protected.POST("/logout", h.Logout)
			protected.GET("/me", h.Me)
			protected.POST("/auth/change-password", h.ChangePassword)

			// Staff
			protected.GET("/staff", h.GetStaff)
			protected.POST("/staff", requireSupportOrAdmin, h.CreateStaff)
			protected.PATCH("/staff/:id/role", requireAdmin, h.UpdateStaffRole)
			protected.POST("/staff/:id/deactivate", requireAdmin, h.DeactivateStaff)
			protected.POST("/staff/:id/activate", requireAdmin, h.ActivateStaff)
			protected.DELETE("/staff/:id", requireAdmin, h.DeleteStaff)

			// Clients
			protected.GET("/clients", h.GetClients)
			protected.POST("/clients", requireSupportOrAdmin, h.CreateClient)
			protected.GET("/clients/:id", h.GetClient)
			protected.PUT("/clients/:id", requireSupportOrAdmin, h.UpdateClient)
			protected.POST("/clients/:id/activate", requireSalesOrAdmin, h.ActivateClient)
			protected.POST("/clients/:id/pause", requireSalesOrAdmin, h.PauseClient)
			protected.POST("/clients/:id/request-deactivate", requireSalesOrAdmin, h.RequestDeactivateClient)
			protected.POST("/clients/:id/approve-deactivate", requireAdmin, h.ApproveDeactivateClient)
			protected.POST("/clients/:id/reject-deactivate", requireAdmin, h.RejectDeactivateClient)

			// Job Requests
			protected.GET("/job-requests", h.GetJobRequests)
			protected.GET("/job-requests/:id", h.GetJobRequest)
			protected.GET("/job-requests/:id/sla", h.GetJobSLA)
			protected.PATCH("/job-requests/:id/assign", requireSupportOrAdmin, h.AssignPICs)
			protected.PATCH("/job-requests/:id/stage1", requireSalesOrAdmin, h.UpdateStage1)
			protected.POST("/job-requests/:id/signed-copy", requireSalesOrAdmin, h.UploadSignedCopy)
			protected.GET("/job-requests/:id/signed-copy/download", h.DownloadSignedCopy)
			protected.GET("/job-requests/:id/agreements", h.GetAgreements)
			protected.POST("/job-requests/:id/agreements", requireSalesOrAdmin, h.UploadAgreement)
			protected.GET("/job-requests/:id/tasks", h.GetTasks)

			// Agreements
			protected.GET("/agreements/:id/download", h.DownloadAgreement)
			protected.POST("/agreements/:id/remarks", requireAdmin, h.AddRemarks)
			protected.POST("/agreements/:id/approve", requireAdmin, h.ApproveAgreement)
			protected.POST("/agreements/:id/reject", requireAdmin, h.RejectAgreement)

			// Tasks
			protected.PATCH("/tasks/:id", requireCS, h.UpdateTask)

			// Projects
			protected.GET("/projects", h.GetProjects)
			protected.GET("/projects/:id/staff", requireAdmin, h.GetProjectStaff)
			protected.POST("/projects/:id/staff", requireAdmin, h.AddProjectStaff)
			protected.DELETE("/projects/:id/staff/:userId", requireAdmin, h.RemoveProjectStaff)

			// Dashboard
			protected.GET("/dashboard", h.GetDashboard)

			// Audit Logs
			protected.GET("/audit-logs", requireAdmin, h.GetAuditLogs)

			// Invoices
			protected.GET("/invoices", requireSalesOrAdmin, h.GetInvoices)
			protected.GET("/invoices/active-clients", requireSalesOrAdmin, h.GetActiveClientsForInvoicing)
			protected.GET("/invoices/admin-overview", requireAdmin, h.GetAdminOverview)
			protected.GET("/invoices/commissions", h.GetCommissions)
			protected.POST("/invoices", requireSalesOrAdmin, h.CreateInvoice)
			protected.POST("/invoices/:id/upload-file", requireSalesOrAdmin, h.UploadInvoiceFile)
			protected.GET("/invoices/:id/download", h.DownloadInvoice)
			protected.POST("/invoices/:id/pay", requireAdmin, h.PayInvoice)
			protected.PATCH("/invoices/:id", requireSalesOrAdmin, h.UpdateInvoice)
		}
	}
}
