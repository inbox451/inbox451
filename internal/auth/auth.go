package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"inbox451/internal/config"
	"inbox451/internal/core"
	"inbox451/internal/logger"
	"inbox451/internal/models"
	"inbox451/internal/storage"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/labstack/echo/v4"
	"github.com/zerodha/simplesessions/stores/postgres/v3"
	"github.com/zerodha/simplesessions/v3"
	"golang.org/x/oauth2"
)

const (
	// UserKey is the key on which the User profile is set on echo handlers.
	UserKey    = "auth_user"
	SessionKey = "auth_session"
)

type OIDCClaim struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Sub           string `json:"sub"`
	Picture       string `json:"picture"`
	Name          string `json:"name"`               // Added Name
	Username      string `json:"preferred_username"` // Often used for username
}

type Auth struct {
	apiTokens map[string]models.Token // Store token value -> Token model
	sync.RWMutex

	core      *core.Core
	oidcCfg   config.OIDCConfig
	oauthCfg  oauth2.Config
	verifier  *oidc.IDTokenVerifier
	provider  *oidc.Provider
	sess      *simplesessions.Manager
	sessStore *postgres.Store
	cb        *Callbacks
	log       *logger.Logger
}

// Callbacks for session manager interaction with Echo context
type Callbacks struct {
	GetUser func(id int) (*models.User, error)
}

// Helper to initialize OIDC provider and config
func initOIDC(a *Auth) {
	if !a.oidcCfg.Enabled {
		return
	}
	ctxOIDC := context.Background()
	a.log.Info("Initializing OIDC provider: %s", a.oidcCfg.ProviderURL)
	provider, err := oidc.NewProvider(ctxOIDC, a.oidcCfg.ProviderURL)
	if err != nil {
		a.oidcCfg.Enabled = false
		a.log.Error("Error initializing OIDC provider, disabling OIDC: %v", err)
		return
	}
	a.provider = provider
	a.oauthCfg = oauth2.Config{
		ClientID:     a.oidcCfg.ClientID,
		ClientSecret: a.oidcCfg.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  a.oidcCfg.RedirectURL,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	a.verifier = provider.Verifier(&oidc.Config{ClientID: a.oidcCfg.ClientID})
	a.log.Info("OIDC Authentication enabled with provider: %s", a.oidcCfg.ProviderURL)
}

// Helper to start session pruning goroutine
func startSessionPruner(ctx context.Context, a *Auth) {
	go func() {
		ticker := time.NewTicker(time.Hour * 1)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				a.log.Debug("Pruning expired sessions")
				err := a.sessStore.Prune()
				if err != nil {
					a.log.Error("Error pruning login sessions: %v", err)
				} else {
					a.log.Debug("Expired session pruning completed")
				}
			case <-ctx.Done():
				a.log.Info("Session pruning goroutine shutting down")
				return
			}
		}
	}()
}

// Helper to start token pruning goroutine
func startTokenPruner(ctx context.Context, a *Auth) {
	go func() {
		ticker := time.NewTicker(time.Hour * 6)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				a.log.Debug("Pruning expired API tokens")
				count, err := a.core.Repository.PruneExpiredTokens(context.Background())
				if err != nil {
					a.log.Error("Error pruning expired tokens: %v", err)
				} else if count > 0 {
					a.log.Info("Pruned %d expired API tokens", count)
				}
			case <-ctx.Done():
				a.log.Info("Token pruning goroutine shutting down")
				return
			}
		}
	}()
}

func New(ctx context.Context, core *core.Core, db *sql.DB, cb *Callbacks) (*Auth, error) {
	a := &Auth{
		core:      core,
		oidcCfg:   core.Config.OIDC,
		cb:        cb,
		log:       core.Logger,
		apiTokens: make(map[string]models.Token),
	}

	initOIDC(a)

	// Initialize session manager
	st, err := postgres.New(postgres.Opt{}, db)
	if err != nil {
		return nil, fmt.Errorf("error creating postgres session store: %w", err)
	}
	a.sessStore = st

	// TODO: Add session store options to config
	// Most of the session options are set in the config file
	// Once everything is working, we should move them to the config file, and probably refactor the entire auth conf
	a.sess = simplesessions.New(simplesessions.Options{
		EnableAutoCreate: false,
		SessionIDLength:  64,
		Cookie: simplesessions.CookieOptions{
			Name:       "inbox451_session",
			IsHTTPOnly: true,
			IsSecure:   false, // Set true if using HTTPS
			MaxAge:     time.Hour * 24 * 7,
			SameSite:   http.SameSiteLaxMode,
			Path:       "/",
		},
	})
	a.sess.UseStore(st)

	// Set cookie hooks for Echo
	a.sess.SetCookieHooks(
		func(name string, r interface{}) (*http.Cookie, error) {
			c := r.(echo.Context)
			cookie, err := c.Cookie(name)
			// Cookie not found is not an error for GetCookie
			if errors.Is(err, http.ErrNoCookie) {
				return nil, nil
			}
			return cookie, err
		},
		func(cookie *http.Cookie, w interface{}) error {
			c := w.(echo.Context)
			c.SetCookie(cookie)
			return nil
		},
	)

	startSessionPruner(ctx, a)
	startTokenPruner(ctx, a)

	return a, nil
}

func (a *Auth) GetAPIToken(tokenValue string) (*models.Token, bool) {
	// 1. Check the cache first (Read Lock)
	a.RLock()
	token, ok := a.apiTokens[tokenValue]
	a.RUnlock()

	if ok {
		// Cache hit: Validate expiration
		if token.ExpiresAt.Valid && token.ExpiresAt.Time.Before(time.Now()) {
			a.log.Info("API token from cache expired: %s (ID: %d, expired_at: %s)", token.Name, token.ID, token.ExpiresAt.Time)
			// TODO: Understand if we should remove expired tokens from the cache
			// We could remove the expired token from the cache here if desired but i'm not sure
			// if this a good idea since users coule be hammering the API with expired tokens and we would then hit
			// the DB for every request. So we just log it and return nil.
			// a.Lock()
			// delete(a.apiTokens, tokenValue)
			// a.Unlock()
			return nil, false // Token found but expired
		}

		// Cache hit and token is valid: Update last used and return
		go a.updateTokenLastUsedAsync(token.ID) // Update last used time
		a.log.Debug("API token cache hit: %s (ID: %d)", token.Name, token.ID)
		return &token, true
	}

	// 2. Cache miss: Fetch from database
	a.log.Debug("API token cache miss, fetching from DB: %s", tokenValue)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Add timeout for DB query
	defer cancel()

	dbToken, err := a.core.Repository.GetTokenByValue(ctx, tokenValue)
	if err != nil {
		// If token is not found in DB, it's invalid
		if errors.Is(err, storage.ErrNotFound) {
			a.log.Warn("API token not found in DB: %s", tokenValue)
			return nil, false // Token not found
		}
		// Handle other potential database errors
		a.log.Error("Error fetching token from DB: %v", err)
		return nil, false // Treat DB error as authentication failure for safety
	}

	// 3. Token found in DB: Validate expiration
	if dbToken.ExpiresAt.Valid && dbToken.ExpiresAt.Time.Before(time.Now()) {
		a.log.Info("API token fetched from DB is expired: %s (ID: %d, expired_at: %s)", dbToken.Name, dbToken.ID, dbToken.ExpiresAt.Time)
		// Don't cache the expired token
		return nil, false // Token found but expired
	}

	// 4. Token found in DB and is valid: Cache it (Write Lock)
	a.Lock()
	a.apiTokens[tokenValue] = *dbToken
	a.Unlock()
	a.log.Info("API token fetched from DB and cached: %s (ID: %d)", dbToken.Name, dbToken.ID)

	// Update last used and return success
	go a.updateTokenLastUsedAsync(dbToken.ID)
	return dbToken, true
}

// updateTokenLastUsedAsync updates the last_used_at timestamp asynchronously.
func (a *Auth) updateTokenLastUsedAsync(tokenID int) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := a.core.Repository.UpdateTokenLastUsed(ctx, tokenID); err != nil {
		a.log.Error("Failed to update last used time for token ID %d: %v", tokenID, err)
	}
}

// GetOIDCAuthURL returns the OIDC provider's auth URL to redirect to.
func (a *Auth) GetOIDCAuthURL(state, nonce string) string {
	if !a.oidcCfg.Enabled {
		return ""
	}
	return a.oauthCfg.AuthCodeURL(state, oidc.Nonce(nonce))
}

// ExchangeOIDCToken takes an OIDC authorization code, validates it, and returns the ID token and claims.
func (a *Auth) ExchangeOIDCToken(code, nonce string) (string, OIDCClaim, error) {
	if !a.oidcCfg.Enabled {
		return "", OIDCClaim{}, errors.New("OIDC is not enabled")
	}

	ctx := context.Background()
	oauth2Token, err := a.oauthCfg.Exchange(ctx, code)
	if err != nil {
		return "", OIDCClaim{}, echo.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf("failed to exchange token: %v", err))
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return "", OIDCClaim{}, echo.NewHTTPError(http.StatusUnauthorized, "no id_token field in oauth2 token")
	}

	idToken, err := a.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return "", OIDCClaim{}, echo.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf("failed to verify ID Token: %v", err))
	}

	// Verify nonce
	if idToken.Nonce != nonce {
		return "", OIDCClaim{}, echo.NewHTTPError(http.StatusUnauthorized, "nonce did not match")
	}

	var claims OIDCClaim
	if err := idToken.Claims(&claims); err != nil {
		return "", OIDCClaim{}, fmt.Errorf("error getting claims: %w", err)
	}

	// Fallback to UserInfo endpoint if email is missing
	if claims.Email == "" {
		userInfo, err := a.provider.UserInfo(ctx, oauth2.StaticTokenSource(oauth2Token))
		if err != nil {
			return "", OIDCClaim{}, fmt.Errorf("error fetching user info from OIDC: %w", err)
		}
		if err := userInfo.Claims(&claims); err != nil {
			return "", OIDCClaim{}, fmt.Errorf("error parsing user info claims: %w", err)
		}
	}

	if claims.Email == "" {
		return "", OIDCClaim{}, errors.New("email claim missing from OIDC token and userinfo")
	}

	return rawIDToken, claims, nil
}

// Helper for API token authentication
func (a *Auth) authenticateAPIToken(c echo.Context) (*models.User, error) {
	authHeader := strings.TrimSpace(c.Request().Header.Get("x-api-key"))
	if authHeader == "" {
		return nil, nil // Not an API token request
	}
	token, ok := a.GetAPIToken(authHeader)
	if !ok {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid API token")
	}
	user, err := a.cb.GetUser(token.UserID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, echo.NewHTTPError(http.StatusUnauthorized, "User associated with API token not found")
		}
		a.log.Error("Error fetching user for API token %d: %v", token.ID, err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Error authenticating API token")
	}
	if user.Status != "active" {
		return nil, echo.NewHTTPError(http.StatusForbidden, "User account is not active")
	}
	return user, nil
}

// Helper for session authentication
func (a *Auth) authenticateSession(c echo.Context) (*models.User, *simplesessions.Session, error) {
	sess, user, err := a.validateSession(c)
	if err != nil {
		return nil, nil, err
	}
	if user.Status != "active" {
		_ = sess.Destroy()
		return nil, nil, echo.NewHTTPError(http.StatusForbidden, "User account is not active")
	}
	return user, sess, nil
}

// Middleware authenticates requests via API token or session cookie.
func (a *Auth) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Try API token authentication
		if user, err := a.authenticateAPIToken(c); err != nil || user != nil {
			if err != nil {
				return err
			}
			c.Set(UserKey, *user)
			return next(c)
		}

		// Try session authentication
		user, sess, err := a.authenticateSession(c)
		if err == nil && user != nil {
			c.Set(UserKey, *user)
			c.Set(SessionKey, sess)
			return next(c)
		}
		if err != nil && !errors.Is(err, simplesessions.ErrInvalidSession) && !errors.Is(err, http.ErrNoCookie) {
			a.log.Error("Session validation error: %v", err)
			return echo.NewHTTPError(http.StatusUnauthorized, "Session error")
		}

		// No valid authentication method found
		return echo.NewHTTPError(http.StatusUnauthorized, "Authentication required")
	}
}

// validateSession checks for a valid session cookie and fetches the user.
func (a *Auth) validateSession(c echo.Context) (*simplesessions.Session, *models.User, error) {
	sess, err := a.sess.Acquire(c.Request().Context(), c, c)
	if err != nil {
		// Distinguish between "no session" and other errors
		if errors.Is(err, simplesessions.ErrInvalidSession) {
			return nil, nil, simplesessions.ErrInvalidSession // No valid session found
		}
		a.log.Error("Error acquiring session: %v", err)
		return nil, nil, fmt.Errorf("error acquiring session: %w", err) // Other session store error
	}

	userIDVal, err := sess.Get("user_id")
	if err != nil {
		// If GetInt returns an error or userID is invalid, the session is likely corrupt or empty
		a.log.Warn("Error getting user_id from session %s: %v", sess.ID(), err)
		_ = sess.Destroy() // Destroy potentially invalid session
		return nil, nil, simplesessions.ErrInvalidSession
	}

	userID, err := a.sessStore.Int(userIDVal, nil)
	if err != nil {
		// User linked to session not found in DB, invalidate session
		if errors.Is(err, storage.ErrNotFound) {
			a.log.Warn("Invalid user_id '%v' in session %s: %v", userIDVal, sess.ID(), err)
			_ = sess.Destroy()
			return nil, nil, simplesessions.ErrInvalidSession
		}
		// Other DB error fetching user
		a.log.Error("Error fetching user %d for session %s: %v", userID, sess.ID(), err)
		return nil, nil, fmt.Errorf("error fetching session user: %w", err)
	}

	user, err := a.cb.GetUser(userID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			a.log.Warn("User %d not found for session %s: %v", userID, sess.ID(), err)
			_ = sess.Destroy()
			return nil, nil, simplesessions.ErrInvalidSession
		}
		a.log.Error("Error fetching user %d for session %s: %v", userID, sess.ID(), err)
		return nil, nil, fmt.Errorf("error fetching session user: %w", err)
	}

	return sess, user, nil
}

// SaveSession creates and sets a session cookie after successful login.
func (a *Auth) SaveSession(u models.User, c echo.Context) error {
	sess, err := a.sess.NewSession(c, c) // Pass echo.Context for cookie hooks
	if err != nil {
		a.log.Error("Error creating login session: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error creating session")
	}

	if err := sess.Set("user_id", u.ID); err != nil {
		a.log.Error("Error setting user_id in session: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error saving session")
	}

	a.log.Info("Session saved successfully for user ID: %d", u.ID)
	return nil
}

// Logout destroys the current user session.
func (a *Auth) Logout(c echo.Context) error {
	sess, ok := c.Get(SessionKey).(*simplesessions.Session)
	if !ok || sess == nil {
		// If no session exists, act as if logout was successful
		return nil
	}
	err := sess.Destroy()
	if err != nil {
		a.log.Error("Error destroying session: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to logout")
	}
	a.log.Info("User logged out, session destroyed: %s", sess.ID())
	return nil
}

// IsOIDCEnabled checks if OIDC authentication is configured and enabled.
func (a *Auth) IsOIDCEnabled() bool {
	return a.oidcCfg.Enabled && a.provider != nil
}
