package service

import librarysvc "github.com/anglesson/simple-web-server/internal/library/service"

// EbookService is an alias for librarysvc.EbookService for backwards compatibility.
// Use librarysvc.EbookService directly in new code.
type EbookService = librarysvc.EbookService

// NewEbookService is an alias for librarysvc.NewEbookService for backwards compatibility.
// Note: signature changed — now requires an EbookRepository parameter.
var NewEbookService = librarysvc.NewEbookService
