package ports

type UUIDGeneratorPort interface {
	GenerateUUID() string
}
