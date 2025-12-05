package validator

import (
	"errors"
	"log"
	"reflect"
	"strings"

	"github.com/go-playground/locales/pt_BR"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	ptBR_translations "github.com/go-playground/validator/v10/translations/pt_BR"
)

var (
	validate *validator.Validate
	trans    ut.Translator
)

func init() {
	validate = validator.New()

	// Registre a função para obter o nome da tag "json"
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		// Tenta obter o nome da tag "json"
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		// Se a tag for "-", o campo deve ser ignorado
		if name == "-" {
			return ""
		}

		// Retorna o nome do campo JSON
		return name
	})

	ptBR := pt_BR.New()
	uni := ut.New(ptBR, ptBR)

	var found bool
	trans, found = uni.GetTranslator("pt_BR")
	if !found {
		log.Fatal("translator not found")
	}

	if err := ptBR_translations.RegisterDefaultTranslations(validate, trans); err != nil {
		log.Fatal(err)
	}

	registerCustomTranslations()
}

func registerCustomTranslations() {
	// Override for 'required'
	_ = validate.RegisterTranslation("required", trans, func(ut ut.Translator) error {
		return ut.Add("required", "O campo {0} é obrigatório", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("required", fe.Field())
		return t
	})

	// Override for 'max'
	_ = validate.RegisterTranslation("max", trans, func(ut ut.Translator) error {
		return ut.Add("max", "O campo {0} deve ter no máximo {1} caracteres", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("max", fe.Field(), fe.Param())
		return t
	})

	// Override for 'ltfield'
	_ = validate.RegisterTranslation("ltfield", trans, func(ut ut.Translator) error {
		return ut.Add("ltfield", "O campo {0} deve ser menor que o campo {1}", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("ltfield", fe.Field(), fe.Param())
		return t
	})
}

// Validate validates the given struct and returns translated error messages.
func Validate(s any) error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return err
	}

	var errorMessages []string
	for _, e := range validationErrors {
		errorMessages = append(errorMessages, e.Translate(trans))
	}

	return errors.New(strings.Join(errorMessages, ", "))
}
