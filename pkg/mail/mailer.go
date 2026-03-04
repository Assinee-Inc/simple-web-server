package mail

type Mailer interface {
	From(name string, email string)
	To(email string)
	Subject(subject string)
	Body(body string)
	Send()
}
