package model

// InputCreateCreator holds the data needed to create a new creator account.
type InputCreateCreator struct {
	Name                 string
	SocialName           string
	CPF                  string
	BirthDate            string
	PhoneNumber          string
	Email                string
	Password             string
	PasswordConfirmation string
	TermsAccepted        string
}
