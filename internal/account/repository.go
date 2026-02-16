package account

type AccountRepository interface {
	Create(account *Account) error
	Update(account *Account) error
}
