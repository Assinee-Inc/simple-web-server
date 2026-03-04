package models

import (
	accountmodel "github.com/anglesson/simple-web-server/internal/account/model"
)

// Creator is a type alias for accountmodel.Creator for backwards compatibility
// during module migration. Use accountmodel.Creator directly in new code.
type Creator = accountmodel.Creator

// NewCreator creates a new Creator. Prefer using accountmodel.NewCreator in new code.
var NewCreator = accountmodel.NewCreator
