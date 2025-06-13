package api

import (
	"context"
	"database/sql"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"inbox451/internal/auth"
	"inbox451/internal/models"

	"inbox451/internal/assets"
	"inbox451/internal/core"
	"inbox451/internal/middleware"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
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
	auth *auth.Auth
}

func NewServer(ctx context.Context, core *core.Core, db *sql.DB) *Server {
	e := echo.New()
	e.HideBanner = true
	s := &Server{
		core: core,
		echo: e,
	}

	// Add timeout middleware with a 30-second timeout
	e.Use(middleware.TimeoutMiddleware(30 * time.Second))

	// Set custom validator
	e.Validator = &CustomValidator{validator: validator.New()}

	// Add middleware
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000"}, // Specific origins instead of "*"
		AllowMethods:     []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true, // Required when sending cookies/credentials
		MaxAge:           86400,
	}))
	e.Use(echomiddleware.RequestID())
	e.Use(echomiddleware.Secure())

	// Set custom error handler
	e.HTTPErrorHandler = s.errorHandler

	// Initialize Auth module
	authCallbacks := &auth.Callbacks{
		GetUser: func(id int) (*models.User, error) {
			return s.core.UserService.Get(context.Background(), id)
		},
	}
	authModule, err := auth.New(ctx, core, db, authCallbacks)
	if err != nil {
		core.Logger.Fatal("Failed to initialize auth module: %v", err)
	}
	s.auth = authModule

	// API routes
	api := e.Group("/api")
	s.routes(api)

	// Serve frontend assets
	e.GET("/*", s.assetHandler)

	return s
}

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

// assetHandler serves frontend assets and index.html fallback
func (s *Server) assetHandler(c echo.Context) error {
	path := c.Param("*")
	if path == "" || path == "/" {
		path = "index.html"
	}

	if path[0] == '/' {
		path = path[1:]
	}

	s.core.Logger.Info("Attempting to serve: %s", path)

	// Try to read the file
	data, err := assets.FS.Read(path)
	if err != nil {
		s.core.Logger.Error("Failed to read file %s: %v", path, err)
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
}

func (s *Server) ListenAndServe() error {
	return s.echo.Start(s.core.Config.Server.HTTP.Port)
}

// Add Shutdown method to Server struct
func (s *Server) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}
