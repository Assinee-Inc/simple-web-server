package mail

import (
	"log"

	"github.com/wneessen/go-mail"
)

type GoMailMailer struct {
	client *mail.Client
	msg    *mail.Msg
}

func NewGoMailer(host string, port int, username, password string) *GoMailMailer {
	c, err := mail.NewClient(host, mail.WithPort(port), mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(username), mail.WithPassword(password))

	if err != nil {
		log.Printf("failed to create mail client: %s", err)
		return nil
	}

	return &GoMailMailer{
		client: c,
		msg:    mail.NewMsg(),
	}
}

func (m *GoMailMailer) From(name string, email string) {
	if err := m.msg.FromFormat(name, email); err != nil {
		log.Printf("failed to set From address: %s", err)
		return
	}
}

func (m *GoMailMailer) To(email string) {
	if err := m.msg.To(email); err != nil {
		log.Printf("failed to set To address: %s", err)
		return
	}
}

func (m *GoMailMailer) Subject(subject string) {
	m.msg.Subject(subject)
}

func (m *GoMailMailer) Body(body string) {
	m.msg.SetBodyString(mail.TypeTextHTML, body)
}

func (m *GoMailMailer) Send() {
	if err := m.client.DialAndSend(m.msg); err != nil {
		log.Printf("failed to send mail: %s", err)
		return
	}

	m.msg.Reset()
}
