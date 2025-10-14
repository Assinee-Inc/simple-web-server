package models

type EbookRequest struct {
	Title            string  `validate:"required,min=5,max=120" json:"title"`
	Description      string  `validate:"required,max=120" json:"description"`
	SalesPage        string  `validate:"required" json:"sales_page"`
	Value            float64 `validate:"required,gt=0" json:"value"`
	PromotionalValue float64 `json:"promotional_value"`
	Status           bool    `json:"status"`
}

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

// User Service Types
type InputCreateUser struct {
	Username             string
	Email                string
	Password             string
	PasswordConfirmation string
}

type InputLogin struct {
	Email    string
	Password string
}

// Creator Service Types
type InputCreateCreator struct {
	Name                 string `json:"name"`
	CPF                  string `json:"cpf"`
	BirthDate            string `json:"birthDate"`
	PhoneNumber          string `json:"phoneNumber"`
	Email                string `json:"email"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"passwordConfirmation"`
	TermsAccepted        string `json:"termsAccepted"`
}
