package repository

import accountrepo "github.com/anglesson/simple-web-server/internal/account/repository"

// CreatorRepository is a type alias for accountrepo.CreatorRepository for backwards compatibility
// during module migration. Use accountrepo.CreatorRepository directly in new code.
type CreatorRepository = accountrepo.CreatorRepository
