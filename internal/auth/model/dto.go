package model

// InputCreateUser holds the data needed to create a new user account.
type InputCreateUser struct {
	Username             string
	Email                string
	Password             string
	PasswordConfirmation string
}

// InputLogin holds the data needed to authenticate a user.
type InputLogin struct {
	Email    string
	Password string
}
