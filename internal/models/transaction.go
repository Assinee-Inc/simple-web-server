package models

import salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"

// SplitType is a type alias for salesmodel.SplitType for backwards compatibility.
type SplitType = salesmodel.SplitType

// TransactionStatus is a type alias for salesmodel.TransactionStatus for backwards compatibility.
type TransactionStatus = salesmodel.TransactionStatus

// Transaction is a type alias for salesmodel.Transaction for backwards compatibility.
type Transaction = salesmodel.Transaction

const (
	SplitTypePercentage  = salesmodel.SplitTypePercentage
	SplitTypeFixedAmount = salesmodel.SplitTypeFixedAmount

	TransactionStatusPending   = salesmodel.TransactionStatusPending
	TransactionStatusCompleted = salesmodel.TransactionStatusCompleted
	TransactionStatusFailed    = salesmodel.TransactionStatusFailed
)

var NewTransaction = salesmodel.NewTransaction
