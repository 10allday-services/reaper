package events

import (
	"bytes"
	"fmt"
	"net/mail"
	"net/smtp"

	"github.com/mostlygeek/reaper/reapable"
)

type Mailer struct {
	conf *SMTPConfig
}

type HTTPConfig struct {
	TokenSecret string
	ApiURL      string
	Listen      string
	Token       string
	Action      string
}

type SMTPConfig struct {
	HTTPConfig
	Enabled  bool
	Host     string
	Port     int
	AuthType string
	Username string
	Password string
	From     FromAddress
}

func (s *SMTPConfig) String() string {
	return fmt.Sprintf("%s:%d auth type:%s, creds: %s:%s",
		s.Host,
		s.Port,
		s.AuthType,
		s.Username,
		s.Password)
}
func (s *SMTPConfig) Addr() string {
	if s.Port == 0 {
		// friends don't let friend's smtp over port 25
		return fmt.Sprintf("%s:%d", s.Host, 587)
	}
	// default
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// Auth creates the appropriate smtp.Auth from the configured AuthType
func (s *SMTPConfig) Auth() smtp.Auth {
	switch s.AuthType {
	case "md5":
		return s.CRAMMD5Auth()
	case "plain":
		return s.PlainAuth()
	default:
		return nil
	}
}

func (s *SMTPConfig) CRAMMD5Auth() smtp.Auth {
	return smtp.CRAMMD5Auth(s.Username, s.Password)
}

func (s *SMTPConfig) PlainAuth() smtp.Auth {
	return smtp.PlainAuth("", s.Username, s.Password, s.Host)
}

type FromAddress struct {
	mail.Address
}

func (f *FromAddress) UnmarshalText(text []byte) error {
	a, err := mail.ParseAddress(string(text))
	if err != nil {
		return err
	}

	f.Address = *a
	return nil
}

func NewMailer(conf *SMTPConfig) *Mailer {
	return &Mailer{conf}
}

// methods to conform to EventReporter interface
func (m *Mailer) NewEvent(title string, text string, fields map[string]string, tags []string) error {
	return nil
}
func (m *Mailer) NewStatistic(name string, value float64, tags []string) error {
	return nil
}
func (m *Mailer) NewCountStatistic(name string, tags []string) error {
	return nil
}

// TODO: figure out how to goroutine this
func (m *Mailer) NewReapableEvent(r Reapable) error {
	addr, subject, body, err := r.ReapableEventEmail()
	if err != nil {
		// if this is an unowned error we don't pass it up
		switch t := err.(type) {
		case reapable.UnownedError:
			log.Error(t.Error())
			return nil
		default:
		}
		return err
	}
	return m.Send(addr, subject, body)
}

// Send an HTML email
func (m *Mailer) Send(to mail.Address, subject, htmlBody string) error {

	buf := bytes.NewBuffer(nil)
	buf.WriteString("From: " + m.conf.From.Address.String() + "\n")
	buf.WriteString("To: " + to.String() + "\n")
	buf.WriteString("Subject: " + subject + "\n")
	buf.WriteString("MIME-Version: 1.0\n")
	buf.WriteString("Content-Type: text/html; charset=utf-8\n\n")
	buf.WriteString(htmlBody)
	buf.WriteString("\n")

	log.Debug("Sending email to: \"%s\", from: \"%s\", subject: \"%s\"",
		to.String(),
		m.conf.From.Address.String(),
		subject)

	return smtp.SendMail(
		m.conf.Addr(),
		m.conf.Auth(),
		m.conf.From.Address.Address,
		[]string{to.Address},
		buf.Bytes(),
	)
}
