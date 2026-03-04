package model

type Login struct {
	HashedPassword string
	SessionToken   string
	CSRFToken      string
}
