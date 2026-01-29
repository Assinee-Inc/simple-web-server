package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

func Translate(err error, obj any) map[string]string {
	errors := make(map[string]string)

	validErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return nil
	}

	for _, f := range validErrs {
		errors[f.Field()] = getMsg(f, obj)
	}
	return errors
}

func getMsg(f validator.FieldError, obj any) string {
	switch f.Tag() {
	case "required":
		return fmt.Sprintf("O campo '%s' é obrigatório", f.Field())
	case "max":
		return fmt.Sprintf("O tamanho máximo é %s caracteres", f.Param())
	case "min":
		return fmt.Sprintf("O tamanho mínimo é %s caracteres", f.Param())
	case "gte":
		return fmt.Sprintf("O valor deve ser maior ou igual a %s", f.Param())
	case "lte":
		return fmt.Sprintf("O valor deve ser menor ou igual a %s", f.Param())
	case "ltfield":
		paramJSONName := getJSONTagName(obj, f.Param())
		fieldJSONName := getJSONTagName(obj, f.Field())
		return fmt.Sprintf("O campo '%s' deve ser menor que o campo '%s'", fieldJSONName, paramJSONName)
	default:
		return "Campo inválido"
	}
}

// Função auxiliar para buscar a tag JSON de um campo pelo nome da Struct
func getJSONTagName(obj any, structFieldName string) string {
	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	field, ok := t.FieldByName(structFieldName)
	if !ok {
		return structFieldName // fallback
	}
	return strings.Split(field.Tag.Get("json"), ",")[0]
}

func NewValidator() *validator.Validate {
	v := validator.New()

	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return v
}
