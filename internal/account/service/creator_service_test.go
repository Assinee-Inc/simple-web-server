package service_test

import (
	"testing"
	"time"

	accountmodel "github.com/Assinee-Inc/docffy/internal/account/model"
	accountsvc "github.com/Assinee-Inc/docffy/internal/account/service"
	authmodel "github.com/Assinee-Inc/docffy/internal/auth/model"
	"github.com/Assinee-Inc/docffy/internal/mocks"
	"github.com/Assinee-Inc/docffy/pkg/gov"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newCreatorServiceSUT() (
	accountsvc.CreatorService,
	*mocks.MockCreatorRepository,
	*mocks.MockRFService,
	*mocks.MockUserService,
	*mocks.MockSubscriptionService,
	*mocks.MockPaymentGateway,
) {
	creatorRepo := new(mocks.MockCreatorRepository)
	rfSvc := new(mocks.MockRFService)
	userSvc := new(mocks.MockUserService)
	subSvc := new(mocks.MockSubscriptionService)
	payGw := new(mocks.MockPaymentGateway)

	svc := accountsvc.NewCreatorService(creatorRepo, rfSvc, userSvc, subSvc, payGw)
	return svc, creatorRepo, rfSvc, userSvc, subSvc, payGw
}

func TestCreatorService_OnboardCreator_Success(t *testing.T) {
	svc, creatorRepo, rfSvc, userSvc, subSvc, payGw := newCreatorServiceSUT()

	now := time.Now()
	authUser := &authmodel.User{
		PublicID:        "usr_10",
		Username:        "Maria Silva",
		Email:           "maria@example.com",
		GoogleID:        "google-10",
		EmailVerifiedAt: &now,
		TermsAcceptedAt: nil,
	}
	authUser.ID = 10

	input := accountsvc.InputOnboardingCreator{
		Name:          "Maria da Silva",
		SocialName:    "Maria Silva",
		CPF:           "52998224725", // CPF válido (checksum)
		BirthDate:     "15/05/1990",
		PhoneNumber:   "(11) 98765-4321",
		TermsAccepted: true,
	}

	userSvc.On("FindByID", uint(10)).Return(authUser)
	creatorRepo.On("FindCreatorByUserID", uint(10)).Return(nil, nil)
	creatorRepo.On("FindByCPF", "52998224725").Return(nil, nil)
	rfSvc.On("ConsultaCPF", "Maria da Silva", "52998224725", "15/05/1990").Return(&gov.ReceitaFederalResponse{
		Status: true,
		Result: gov.ConsultaData{
			NomeDaPF: "MARIA DA SILVA",
		},
	}, nil)

	userSvc.On("Update", mock.MatchedBy(func(u *authmodel.User) bool {
		return u.ID == 10 && u.TermsAcceptedAt != nil
	})).Return(nil)

	creatorRepo.On("Create", mock.MatchedBy(func(c *accountmodel.Creator) bool {
		c.ID = 55
		c.PublicID = "crt_55"
		return c.UserID == 10 &&
			c.Email == "maria@example.com" &&
			c.CPF == "52998224725" &&
			c.Phone == "11987654321"
	})).Return(nil)

	payGw.On("CreateCustomer", "maria@example.com", "MARIA DA SILVA").Return("cus_123", nil)
	subSvc.On("CreateSubscription", uint(10), "default_plan").Return(uint(1), nil)
	subSvc.On("ActivateSubscription", uint(1), "cus_123", "").Return(nil)

	creator, err := svc.OnboardCreator(10, input)
	require.NoError(t, err)
	require.NotNil(t, creator)
	assert.Equal(t, uint(55), creator.ID)
	assert.Equal(t, uint(10), creator.UserID)
	assert.Equal(t, "maria@example.com", creator.Email)
	assert.Equal(t, "52998224725", creator.CPF)
	assert.NotNil(t, authUser.TermsAcceptedAt)

	userSvc.AssertExpectations(t)
	creatorRepo.AssertExpectations(t)
}

func TestCreatorService_OnboardCreator_ValidationsAndConflicts(t *testing.T) {
	now := time.Now()
	authUser := &authmodel.User{
		PublicID:        "usr_20",
		Username:        "Teste",
		Email:           "teste@example.com",
		EmailVerifiedAt: &now,
	}
	authUser.ID = 20

	t.Run("rejeita menor de 18 anos", func(t *testing.T) {
		svc, _, _, _, _, _ := newCreatorServiceSUT()
		underageDate := time.Now().AddDate(-16, 0, 0).Format("02/01/2006")
		creator, err := svc.OnboardCreator(20, accountsvc.InputOnboardingCreator{
			Name:          "Menor de Idade",
			CPF:           "52998224725",
			BirthDate:     underageDate,
			PhoneNumber:   "11987654321",
			TermsAccepted: true,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "18 anos")
		assert.Nil(t, creator)
	})

	t.Run("rejeita CPF com checksum inválido", func(t *testing.T) {
		svc, _, _, _, _, _ := newCreatorServiceSUT()
		creator, err := svc.OnboardCreator(20, accountsvc.InputOnboardingCreator{
			Name:          "CPF Inválido",
			CPF:           "11111111111",
			BirthDate:     "15/05/1990",
			PhoneNumber:   "11987654321",
			TermsAccepted: true,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CPF inválido")
		assert.Nil(t, creator)
	})

	t.Run("rejeita quando termos de uso não foram aceitos", func(t *testing.T) {
		svc, _, _, _, _, _ := newCreatorServiceSUT()
		creator, err := svc.OnboardCreator(20, accountsvc.InputOnboardingCreator{
			Name:          "Sem Termos",
			CPF:           "52998224725",
			BirthDate:     "15/05/1990",
			PhoneNumber:   "11987654321",
			TermsAccepted: false,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Termos")
		assert.Nil(t, creator)
	})

	t.Run("retorna erro de conflito quando CPF já está cadastrado", func(t *testing.T) {
		svc, creatorRepo, _, userSvc, _, _ := newCreatorServiceSUT()
		userSvc.On("FindByID", uint(20)).Return(authUser)
		creatorRepo.On("FindCreatorByUserID", uint(20)).Return(nil, nil)
		creatorRepo.On("FindByCPF", "52998224725").Return(&accountmodel.Creator{CPF: "52998224725"}, nil)

		creator, err := svc.OnboardCreator(20, accountsvc.InputOnboardingCreator{
			Name:          "Maria da Silva",
			CPF:           "52998224725",
			BirthDate:     "15/05/1990",
			PhoneNumber:   "11987654321",
			TermsAccepted: true,
		})
		assert.ErrorIs(t, err, accountsvc.ErrCreatorAlreadyExists)
		assert.Nil(t, creator)
		userSvc.AssertNotCalled(t, "Update")
	})
}
