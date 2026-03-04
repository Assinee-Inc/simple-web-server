package repository

import libraryrepo "github.com/anglesson/simple-web-server/internal/library/repository"

// FileQuery is an alias for libraryrepo.FileQuery for backwards compatibility.
// Use libraryrepo.FileQuery directly in new code.
type FileQuery = libraryrepo.FileQuery

// FileRepository is an alias for libraryrepo.FileRepository for backwards compatibility.
// Use libraryrepo.FileRepository directly in new code.
type FileRepository = libraryrepo.FileRepository

// GormFileRepository is an alias for libraryrepo.GormFileRepository for backwards compatibility.
// Use libraryrepo.GormFileRepository directly in new code.
type GormFileRepository = libraryrepo.GormFileRepository

// NewGormFileRepository is an alias for libraryrepo.NewGormFileRepository for backwards compatibility.
var NewGormFileRepository = libraryrepo.NewGormFileRepository
