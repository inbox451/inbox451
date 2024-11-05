package api

import (
	"context"
	"mercury/internal/core"
	"net/http"
	"strings"
	"time"

	"mercury/internal/assets"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"mime"
	"path/filepath"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

type Server struct {
	core *core.Core
	echo *echo.Echo
}

func NewServer(core *core.Core) *Server {
	e := echo.New()
	s := &Server{
		core: core,
		echo: e,
	}

	// Add timeout middleware with a 30-second timeout
	e.Use(TimeoutMiddleware(30 * time.Second))

	// Set custom validator
	e.Validator = &CustomValidator{validator: validator.New()}

	// Add middleware
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())
	e.Use(middleware.Secure())

	// Set custom error handler
	e.HTTPErrorHandler = s.errorHandler

	// API routes
	api := e.Group("/api")
	s.routes(api)

	// Serve frontend assets
	e.GET("/*", func(c echo.Context) error {
		path := c.Param("*")
		if path == "" || path == "/" {
			path = "index.html"
		}

		if path[0] == '/' {
			path = path[1:]
		}

		core.Logger.Info("Attempting to serve: %s", path)

		// Try to read the file
		data, err := assets.FS.Read(path)
		if err != nil {
			core.Logger.Error("Failed to read file %s: %v", path, err)
			// If the file is not found and it's not an API route, serve index.html
			if !strings.HasPrefix(path, "api/") {
				indexData, err := assets.FS.Read("index.html")
				if err != nil {
					return c.String(http.StatusNotFound, "File not found")
				}
				return c.HTMLBlob(http.StatusOK, indexData)
			}
			return c.String(http.StatusNotFound, "File not found")
		}

		// Determine content type based on file extension
		contentType := mime.TypeByExtension(filepath.Ext(path))
		if contentType == "" {
			contentType = http.DetectContentType(data)
		}

		return c.Blob(http.StatusOK, contentType, data)
	})

	return s
}

// Add the error handler method
func (s *Server) errorHandler(err error, c echo.Context) {
	if he, ok := err.(*echo.HTTPError); ok {
		if he.Internal != nil {
			if herr, ok := he.Internal.(*echo.HTTPError); ok {
				he = herr
			}
		}
		if err := c.JSON(he.Code, he.Message); err != nil {
			s.core.Logger.Error("Failed to send error response: %v", err)
		}
		return
	}

	if err := s.core.HandleError(err, 500); err != nil {
		if err := c.JSON(500, err); err != nil {
			s.core.Logger.Error("Failed to send error response: %v", err)
		}
	}
}

func (s *Server) routes(api *echo.Group) {
	// Account routes
	api.POST("/accounts", s.createAccount)
	api.GET("/accounts", s.getAccounts)
	api.GET("/accounts/:id", s.getAccount)
	api.PUT("/accounts/:id", s.updateAccount)
	api.DELETE("/accounts/:id", s.deleteAccount)

	// Inbox routes
	api.POST("/accounts/:accountId/inboxes", s.createInbox)
	api.GET("/accounts/:accountId/inboxes", s.getInboxes)
	api.GET("/accounts/:accountId/inboxes/:inboxId", s.getInbox)
	api.PUT("/accounts/:accountId/inboxes/:inboxId", s.updateInbox)
	api.DELETE("/accounts/:accountId/inboxes/:inboxId", s.deleteInbox)

	// Rule routes
	api.POST("/accounts/:accountId/inboxes/:inboxId/rules", s.createRule)
	api.GET("/accounts/:accountId/inboxes/:inboxId/rules", s.getRules)
	api.GET("/accounts/:accountId/inboxes/:inboxId/rules/:ruleId", s.getRule)
	api.PUT("/accounts/:accountId/inboxes/:inboxId/rules/:ruleId", s.updateRule)
	api.DELETE("/accounts/:accountId/inboxes/:inboxId/rules/:ruleId", s.deleteRule)

	// Message routes
	api.GET("/accounts/:accountId/inboxes/:inboxId/messages", s.getMessages)
}

func (s *Server) ListenAndServe() error {
	return s.echo.Start(s.core.Config.Server.HTTP.Port)
}

// Add Shutdown method to Server struct
func (s *Server) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}
