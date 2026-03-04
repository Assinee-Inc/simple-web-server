// Package service provides backward-compatible shims for types that have been
// migrated to their respective domain modules. Import from the domain modules directly
// in new code.
package service

import authsvc "github.com/anglesson/simple-web-server/internal/auth/service"

const (
	UserIDKey    = authsvc.UserIDKey
	UserEmailKey = authsvc.UserEmailKey
	CSRFTokenKey = authsvc.CSRFTokenKey
)

// SessionService is a type alias for authsvc.SessionService for backwards compatibility.
// Use authsvc.SessionService in new code.
type SessionService = authsvc.SessionService

// NewSessionService creates a new SessionService. Prefer authsvc.NewSessionService in new code.
var NewSessionService = authsvc.NewSessionService
