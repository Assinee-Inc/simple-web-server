package handler

import libraryhandler "github.com/anglesson/simple-web-server/internal/library/handler"

// EbookHandler is an alias for libraryhandler.EbookHandler for backwards compatibility.
// Use libraryhandler.EbookHandler directly in new code.
type EbookHandler = libraryhandler.EbookHandler

// NewEbookHandler is an alias for libraryhandler.NewEbookHandler for backwards compatibility.
var NewEbookHandler = libraryhandler.NewEbookHandler
