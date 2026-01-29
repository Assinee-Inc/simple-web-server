package uuid

import (
	"fmt"

	"github.com/google/uuid"
)

type IDGenerator interface {
	Generate() string
}

type GoogleUUID struct {
}

func NewGoogleUUID() *GoogleUUID {
	return &GoogleUUID{}
}

func (g *GoogleUUID) Generate() string {
	v7, err := uuid.NewV7()
	if err != nil {
		panic(fmt.Sprintf("erro ao gerar uuid: %s", err.Error()))
	}
	return v7.String()
}
