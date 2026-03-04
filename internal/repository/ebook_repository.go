package repository

import libraryrepo "github.com/anglesson/simple-web-server/internal/library/repository"

// EbookQuery is an alias for libraryrepo.EbookQuery for backwards compatibility.
// Use libraryrepo.EbookQuery directly in new code.
type EbookQuery = libraryrepo.EbookQuery

// EbookRepository is an alias for libraryrepo.EbookRepository for backwards compatibility.
// Use libraryrepo.EbookRepository directly in new code.
type EbookRepository = libraryrepo.EbookRepository

// GormEbookRepository is an alias for libraryrepo.GormEbookRepository for backwards compatibility.
// Use libraryrepo.GormEbookRepository directly in new code.
type GormEbookRepository = libraryrepo.GormEbookRepository

// NewGormEbookRepository is an alias for libraryrepo.NewGormEbookRepository for backwards compatibility.
var NewGormEbookRepository = libraryrepo.NewGormEbookRepository
