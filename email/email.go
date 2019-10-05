package email

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"text/template"
	"time"

	mailgun "github.com/mailgun/mailgun-go/v3"
)

// Email structure which holds the parsed HTML template for the emails
type Email struct {
	From    string
	To      []string
	Subject string
	Text    string
	HTML    []byte
	Data    interface{}
	Images  []string
}

// SendEmail will handle all the generation of automaticly sent emails from this API/project.
// Templates can be found end customized in the "emails" folder under "extras" package
func (em *Email) sendEmail(html string) error {
	mg, err := mailgun.NewMailgunFromEnv()
	if err != nil {
		return err
	}

	message := mg.NewMessage(
		em.From,
		em.Subject,
		em.Text,
		em.To...,
	)

	err = em.readTemplate(html)
	if err != nil {
		return err
	}

	message.SetHtml(string(em.HTML))

	for _, image := range em.Images {
		message.AddInline(fmt.Sprintf("email/templates/images/%s", image))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	_, _, err = mg.Send(ctx, message)

	return err
}

func (em *Email) readTemplate(html string) error {
	t, err := template.ParseFiles(fmt.Sprintf("email/templates/%s", html))
	if err != nil {
		log.Println(err)
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, em.Data); err != nil {
		return err
	}

	em.HTML = buf.Bytes()
	return nil
}
