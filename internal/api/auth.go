package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"inbox451/internal/auth"
	"inbox451/internal/core"
	"inbox451/internal/models"
	"inbox451/internal/storage"

	"github.com/labstack/echo/v4"
)

type loginRequest struct {
	Username string `json:"username" form:"username" validate:"required"`
	Password string `json:"password" form:"password" validate:"required"` // Add validation
}

// login handles user login via username and password.
func (s *Server) login(c echo.Context) error {
	req := new(loginRequest)
	if err := c.Bind(req); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}

	// Validate input format first
	if err := c.Validate(req); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}

	ctx := c.Request().Context()
	// TODO: Check if the login we would attempt to login via e-mail and username or simply username
	user, err := s.core.UserService.LoginWithPassword(ctx, req.Username, req.Password)
	if err != nil {
		// HandleError will determine the correct HTTP status code based on the error type
		return s.core.HandleError(err, http.StatusUnauthorized) // Default to Unauthorized for login failures
	}

	// Login successful, save session
	if err := s.auth.SaveSession(*user, c); err != nil {
		// This is an internal server error (session saving failed)
		return s.core.HandleError(err, http.StatusInternalServerError)
	}

	s.core.Logger.Info("User %s (ID: %s) logged in successfully via password.", user.Username, user.ID)
	// Return only non-sensitive user info
	userInfo := map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"name":     user.Name,
		"email":    user.Email,
		"role":     user.Role,
	}
	return c.JSON(http.StatusOK, userInfo)
}

// logout handles user logout by destroying the session.
func (s *Server) logout(c echo.Context) error {
	if err := s.auth.Logout(c); err != nil {
		// HandleError will determine the correct status code
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// profile returns the currently authenticated user's profile.
func (s *Server) profile(c echo.Context) error {
	// The user object is already set by the auth middleware
	user, ok := c.Get(auth.UserKey).(models.User)
	if !ok {
		// This should technically not happen if middleware is applied correctly
		return s.core.HandleError(echo.NewHTTPError(http.StatusInternalServerError, "User not found in context"), http.StatusInternalServerError)
	}

	// Return non-sensitive user info
	userInfo := map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"name":     user.Name,
		"email":    user.Email,
		"role":     user.Role,
	}
	return c.JSON(http.StatusOK, userInfo)
}

// --- OIDC Handlers (Add these if OIDC is enabled) ---

// oidcLogin redirects the user to the OIDC provider.
func (s *Server) oidcLogin(c echo.Context) error {
	if !s.auth.IsOIDCEnabled() {
		return c.String(http.StatusNotImplemented, "OIDC login is not enabled")
	}

	// Generate state and nonce (simple example, enhance for production)
	state, _ := core.GenerateSecureTokenBase64()
	nonce, _ := core.GenerateSecureTokenBase64()

	// Store state/nonce temporarily (e.g., in a short-lived cookie or server-side cache)
	// these cookies don't have anything to do with the user session, if they need to be configurable
	// they should be under the oidc config

	c.SetCookie(&http.Cookie{
		Name:     "oidc_state",
		Value:    state,
		Path:     "/",
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		// Secure: true, // Uncomment if using HTTPS, this should be configurable
	})
	c.SetCookie(&http.Cookie{
		Name:     "oidc_nonce",
		Value:    nonce,
		Path:     "/",
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		// Secure: true, // Uncomment if using HTTPS, this should be configurable
	})

	authURL := s.auth.GetOIDCAuthURL(state, nonce)
	return c.Redirect(http.StatusFound, authURL)
}

// oidcCallback handles the callback from the OIDC provider.
//
// TODO:
// - Handle different claims based on OIDC provider
// - Implement user provisioning if the user does not exist (code is there but commented out)
func (s *Server) oidcCallback(c echo.Context) error {
	if !s.auth.IsOIDCEnabled() {
		return c.String(http.StatusNotImplemented, "OIDC login is not enabled")
	}

	// Retrieve state and nonce from cookies
	stateCookie, err := c.Cookie("oidc_state")
	if err != nil || stateCookie.Value == "" {
		return s.core.HandleError(errors.New("OIDC state cookie missing or empty"), http.StatusBadRequest)
	}
	nonceCookie, err := c.Cookie("oidc_nonce")
	if err != nil || nonceCookie.Value == "" {
		return s.core.HandleError(errors.New("OIDC nonce cookie missing or empty"), http.StatusBadRequest)
	}

	// Clear cookies immediately after retrieving
	stateCookie.MaxAge = -1
	c.SetCookie(stateCookie)
	nonceCookie.MaxAge = -1
	c.SetCookie(nonceCookie)

	// Validate state parameter
	if c.QueryParam("state") != stateCookie.Value {
		return s.core.HandleError(errors.New("OIDC state mismatch"), http.StatusBadRequest)
	}

	// Exchange code for token and verify
	_, claims, err := s.auth.ExchangeOIDCToken(c.QueryParam("code"), nonceCookie.Value)
	if err != nil {
		return s.core.HandleError(fmt.Errorf("OIDC token exchange/verification failed: %w", err), http.StatusUnauthorized)
	}

	// Find user by email claim
	// TODO: Handle different claims based on OIDC provider
	// here we will to have a mapping feature since we already have knowledge that some oidc providers don't return
	// email but username and under a different claim
	ctx := c.Request().Context()
	user, err := s.core.Repository.GetUserByEmail(ctx, claims.Email)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return s.core.HandleError(fmt.Errorf("database error finding user by email: %w", err), http.StatusInternalServerError)
	}

	// TODO: Auto-provisioning logic
	// We should have in the configuration a flag to enable/disable auto-provisioning
	if user == nil {
		// User doesn't exist - should we provision a new one or deny login?
		// For now, let's deny login if user doesn't pre-exist
		s.core.Logger.Warn("OIDC login failed: User with email %s not found.", claims.Email)
		return s.core.HandleError(errors.New("user not registered"), http.StatusForbidden)
		/*
			newUser := &models.User{
				Email:         claims.Email,
				Name:          claims.Name,         // Assuming name is in claims
				Username:      claims.Username,     // Assuming username is in claims, or generate one
				Status:        "active",
				Role:          "user",             // Default role
				PasswordLogin: false,            // Disable password login for OIDC users by default
			}
			// Password should be NULL or a randomly generated unusable hash
			newUser.Password = null.StringFromPtr(nil)

			if err := s.core.UserService.Create(ctx, newUser); err != nil {
				return s.core.HandleError(fmt.Errorf("failed to auto-provision OIDC user: %w", err), http.StatusInternalServerError)
			}
			user = newUser // Use the newly created user
			s.core.Logger.Info("Auto-provisioned new user via OIDC: %s (ID: %d)", user.Email, user.ID)
		*/
	}

	// User exists, check if active
	if user.Status != "active" {
		s.core.Logger.Warn("OIDC login failed: User account %s is inactive.", claims.Email)
		return s.core.HandleError(core.ErrAccountInactive, http.StatusForbidden)
	}

	// Login successful, save session
	if err := s.auth.SaveSession(*user, c); err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}

	s.core.Logger.Info("User %s (ID: %s) logged in successfully via OIDC.", user.Username, user.ID)

	// Redirect to the frontend (e.g., dashboard)
	// Ideally, the original 'next' URL should be part of the state (need to align with bernardo)
	// but for simplicity, redirecting to a fixed path for now.
	return c.Redirect(http.StatusFound, "/") // Adjust redirect path as needed
}
