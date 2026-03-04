package service

import accountsvc "github.com/anglesson/simple-web-server/internal/account/service"

// Type aliases for backwards compatibility during module migration.
// Use accountsvc types directly in new code.

type RequirementPriority = accountsvc.RequirementPriority
type RequirementType = accountsvc.RequirementType
type PendingRequirement = accountsvc.PendingRequirement
type AccountRequirementsStatus = accountsvc.AccountRequirementsStatus

const (
	PriorityPastDue    = accountsvc.PriorityPastDue
	PriorityCurrently  = accountsvc.PriorityCurrently
	PriorityEventually = accountsvc.PriorityEventually
)

const (
	RequirementTypeDocument   = accountsvc.RequirementTypeDocument
	RequirementTypePersonal   = accountsvc.RequirementTypePersonal
	RequirementTypeCompliance = accountsvc.RequirementTypeCompliance
	RequirementTypeOther      = accountsvc.RequirementTypeOther
)

// StripeConnectService is a type alias for accountsvc.StripeConnectService for backwards compatibility.
// Use accountsvc.StripeConnectService directly in new code.
type StripeConnectService = accountsvc.StripeConnectService

// NewStripeConnectService creates a new StripeConnectService.
var NewStripeConnectService = accountsvc.NewStripeConnectService
