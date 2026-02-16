package account

type AccountGateway interface {
	CreateSellerAccount(account *Account) (string, error)
}
