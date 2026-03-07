package models

import (
	authmodel "github.com/anglesson/simple-web-server/internal/auth/model"
	librarymodel "github.com/anglesson/simple-web-server/internal/library/model"
	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
)

// InputCreateUser is a type alias for authmodel.InputCreateUser for backwards compatibility.
type InputCreateUser = authmodel.InputCreateUser

// InputLogin is a type alias for authmodel.InputLogin for backwards compatibility.
type InputLogin = authmodel.InputLogin

// EbookRequest is a type alias for librarymodel.EbookRequest for backwards compatibility.
type EbookRequest = librarymodel.EbookRequest

type LoginForm struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type FormErrors map[string]string

type RegisterForm struct {
	Username             string `json:"username"`
	Email                string `json:"email"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password_confirmation"`
	TermsAccepted        string `json:"terms_accepted"`
}

// ClientRequest is a type alias for salesmodel.ClientRequest for backwards compatibility.
type ClientRequest = salesmodel.ClientRequest

// CreateClientInput is a type alias for salesmodel.CreateClientInput for backwards compatibility.
type CreateClientInput = salesmodel.CreateClientInput

// CreateClientOutput is a type alias for salesmodel.CreateClientOutput for backwards compatibility.
type CreateClientOutput = salesmodel.CreateClientOutput

// UpdateClientInput is a type alias for salesmodel.UpdateClientInput for backwards compatibility.
type UpdateClientInput = salesmodel.UpdateClientInput
