package ports

type Validator interface {
	Validate(s any) error
}
