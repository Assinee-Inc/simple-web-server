package service_test

import (
	"reflect"
	"testing"

	authsvc "github.com/Assinee-Inc/docffy/internal/auth/service"
	"github.com/Assinee-Inc/docffy/internal/mocks"
	"github.com/stretchr/testify/assert"
)

func TestEmailService_LegacyMethodsRemoved(t *testing.T) {
	mockMailer := new(mocks.MockMailerSimple)
	svc := authsvc.NewEmailService(mockMailer)
	assert.NotNil(t, svc)

	svcType := reflect.TypeOf(svc)
	_, hasReset := svcType.MethodByName("SendPasswordResetEmail")
	_, hasConfirm := svcType.MethodByName("SendAccountConfirmation")

	assert.False(t, hasReset, "SendPasswordResetEmail deve ter sido removido (FR-002)")
	assert.False(t, hasConfirm, "SendAccountConfirmation deve ter sido removido (FR-003)")
}
