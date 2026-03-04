package service

import librarysvc "github.com/anglesson/simple-web-server/internal/library/service"

// FileService is an alias for librarysvc.FileService for backwards compatibility.
// Use librarysvc.FileService directly in new code.
type FileService = librarysvc.FileService

// NewFileService is an alias for librarysvc.NewFileService for backwards compatibility.
var NewFileService = librarysvc.NewFileService
