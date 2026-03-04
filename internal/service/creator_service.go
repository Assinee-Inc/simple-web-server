package service

import accountsvc "github.com/anglesson/simple-web-server/internal/account/service"

// CreatorService is a type alias for accountsvc.CreatorService for backwards compatibility
// during module migration. Use accountsvc.CreatorService directly in new code.
type CreatorService = accountsvc.CreatorService
