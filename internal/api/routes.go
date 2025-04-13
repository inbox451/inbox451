package api

import (
	"github.com/labstack/echo/v4"
)

func (s *Server) routes(api *echo.Group) {
	// Health check endpoint
	api.GET("/api/health", s.healthCheck)

	// -- Authentication routes
	authGroup := s.echo.Group("/api/auth")
	authGroup.POST("/login", s.login)
	authGroup.POST("/logout", s.logout, s.auth.Middleware)
	if s.auth.IsOIDCEnabled() {
		authGroup.GET("/oidc/login", s.oidcLogin)
		authGroup.GET("/oidc/callback", s.oidcCallback)
	}

	// Protect API routes
	// Apply the authentication middleware to all routes in the API group
	api.Use(s.auth.Middleware)

	// User routes
	api.GET("/users", s.getUsers)
	api.GET("/users/me", s.profile)
	api.GET("/users/:userId", s.getUser)
	api.GET("/users/:userId/projects", s.getProjectsByUser)
	api.POST("/users", s.createUser)
	api.PUT("/users/:userId", s.updateUser)
	api.DELETE("/users/:userId", s.deleteUser)

	// ProjectUser routes
	api.POST("/projects/:projectId/users", s.projectAddUser)
	api.DELETE("/projects/:projectId/users/:userId", s.projectRemoveUser)

	// Project routes
	api.GET("/projects", s.getProjects)
	api.GET("/projects/:projectId", s.getProject)
	api.POST("/projects", s.createProject)
	api.PUT("/projects/:projectId", s.updateProject)
	api.DELETE("/projects/:projectId", s.deleteProject)

	// Token routes
	api.GET("/users/:userId/tokens", s.ListTokensByUser)
	api.GET("/users/:userId/tokens/:tokenId", s.GetTokenByUser)
	api.POST("/users/:userId/tokens", s.CreateTokenForUser)
	api.DELETE("/users/:userId/tokens/:tokenId", s.DeleteTokenByUser)

	// Inbox routes
	api.GET("/projects/:projectId/inboxes", s.getInboxes)
	api.GET("/projects/:projectId/inboxes/:inboxId", s.getInbox)
	api.POST("/projects/:projectId/inboxes", s.createInbox)
	api.PUT("/projects/:projectId/inboxes/:inboxId", s.updateInbox)
	api.DELETE("/projects/:projectId/inboxes/:inboxId", s.deleteInbox)

	// Rule routes
	api.GET("/projects/:projectId/inboxes/:inboxId/rules", s.getRules)
	api.GET("/projects/:projectId/inboxes/:inboxId/rules/:ruleId", s.getRule)
	api.POST("/projects/:projectId/inboxes/:inboxId/rules", s.createRule)
	api.PUT("/projects/:projectId/inboxes/:inboxId/rules/:ruleId", s.updateRule)
	api.DELETE("/projects/:projectId/inboxes/:inboxId/rules/:ruleId", s.deleteRule)

	// Message routes
	api.GET("/projects/:projectId/inboxes/:inboxId/messages", s.getMessages)
	api.GET("/projects/:projectId/inboxes/:inboxId/messages/:messageId", s.getMessage)
	api.PUT("/projects/:projectId/inboxes/:inboxId/messages/:messageId/read", s.markMessageRead)
	api.PUT("/projects/:projectId/inboxes/:inboxId/messages/:messageId/unread", s.markMessageUnread)
	api.DELETE("/projects/:projectId/inboxes/:inboxId/messages/:messageId", s.deleteMessage)

}
