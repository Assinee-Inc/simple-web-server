package handler

import libraryhandler "github.com/anglesson/simple-web-server/internal/library/handler"

// FileHandler is an alias for libraryhandler.FileHandler for backwards compatibility.
// Use libraryhandler.FileHandler directly in new code.
type FileHandler = libraryhandler.FileHandler

// NewFileHandler is an alias for libraryhandler.NewFileHandler for backwards compatibility.
var NewFileHandler = libraryhandler.NewFileHandler
