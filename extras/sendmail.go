package extras

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
	"text/template"

	"github.com/jordan-wright/email"
)

type Email struct {
	HTML []byte
}

func SendEmail(from string, to []string, subject string, html EmailTemplates, data interface{}) error {
	em := Email{}
	e := email.NewEmail()
	e.From = from
	e.To = to
	e.Subject = subject

	err := em.readTemplate(html, data)
	if err != nil {
		return err
	}

	e.Headers.Add("Content-Type", "text/html")
	e.Headers.Add("charset", "utf-8")

	e.HTML = em.HTML

	err = e.SendWithTLS(fmt.Sprintf("%s:%s", os.Getenv("SMTP_HOST"), os.Getenv("SMTP_PORT")), smtp.PlainAuth("", os.Getenv("SMTP_USERNAME"), os.Getenv("SMTP_PASSWORD"), os.Getenv("SMTP_HOST")), &tls.Config{ServerName: os.Getenv("SMTP_HOST")})
	if err != nil {
		return err
	}
	return nil
}

func (em *Email) readTemplate(html EmailTemplates, data interface{}) error {
	t, err := template.ParseFiles(fmt.Sprintf("extras/emails/%s", html))
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return err
	}

	em.HTML = buf.Bytes()
	return nil
}
