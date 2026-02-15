package account

type AccountRepository interface {
	Create(account *Account) error
	Update(account *Account) error
}

type AccountGateway interface {
	CreateSellerAccount(account *Account) (string, error)
}

type Manager struct {
	accountRepo AccountRepository
	stripeSvc   AccountGateway
}

func NewManager(accountRepo AccountRepository, stripeSvc AccountGateway) *Manager {
	return &Manager{
		accountRepo: accountRepo,
		stripeSvc:   stripeSvc,
	}
}

func (m *Manager) CreateAccount(account *Account) error {
	err := m.accountRepo.Create(account)
	if err != nil {
		return InternalError
	}

	connectedAccountID, err := m.stripeSvc.CreateSellerAccount(account)
	if err != nil {
		return StripeIntegrationError
	} else {
		account.ExternalAccountID = connectedAccountID
		err = m.accountRepo.Update(account)
		if err != nil {
			return InternalError
		}
	}

	return nil
}
