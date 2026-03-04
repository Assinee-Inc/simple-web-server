package models

import (
	authmodel "github.com/anglesson/simple-web-server/internal/auth/model"
	librarymodel "github.com/anglesson/simple-web-server/internal/library/model"
)

// InputCreateUser is a type alias for authmodel.InputCreateUser for backwards compatibility.
// Use authmodel.InputCreateUser directly in new code.
type InputCreateUser = authmodel.InputCreateUser

// InputLogin is a type alias for authmodel.InputLogin for backwards compatibility.
// Use authmodel.InputLogin directly in new code.
type InputLogin = authmodel.InputLogin

// EbookRequest is a type alias for librarymodel.EbookRequest for backwards compatibility.
// Use librarymodel.EbookRequest directly in new code.
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

type ClientRequest struct {
	ID        uint   `json:"id"`
	Name      string `validate:"required,min=5,max=120" json:"name"`
	CPF       string `validate:"required,max=120" json:"cpf"`
	Birthdate string `validate:"required"`
	Email     string `validate:"required,email" json:"email"`
	Phone     string `validate:"max=14" json:"phone"`
}

// Client Service Types
type CreateClientInput struct {
	Name         string
	CPF          string
	Phone        string
	BirthDate    string
	Email        string
	EmailCreator string
}

type CreateClientOutput struct {
	ID        int
	Name      string
	CPF       string
	Phone     string
	BirthDate string
	Email     string
	CreatedAt string
	UpdatedAt string
}

type UpdateClientInput struct {
	ID           uint
	Email        string
	Phone        string
	EmailCreator string
}

