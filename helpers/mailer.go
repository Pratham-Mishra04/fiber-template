package helpers

import (
	"bytes"
	"html/template"

	"github.com/Pratham-Mishra04/fiber-template/config"
	"github.com/Pratham-Mishra04/fiber-template/initializers"
	"github.com/go-gomail/gomail"
)

func SendMail(subject string, body string, recipientName string, recipientEmail string, htmlStr string) error {
	htmlContent := body + htmlStr
	m := gomail.NewMessage()
	m.SetHeader("From", config.EMAIL_SENDER)
	m.SetHeader("To", recipientEmail)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlContent)

	d := gomail.NewDialer("smtp.gmail.com", 587, config.EMAIL_SENDER, initializers.CONFIG.GMAIL_KEY)

	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}

func SendTemplateMail(recipientName string, recipientEmail string, subject string, templateName string) {
	var body bytes.Buffer
	path := config.TEMPLATE_DIR + templateName
	t, err := template.ParseFiles(path)
	if err != nil {
		LogDatabaseError("Error while sending Mail", err, "go_routine")
	}

	t.Execute(&body, struct {
		Name string
	}{Name: recipientName})

	if err != nil {
		LogDatabaseError("Error while sending Mail", err, "go_routine")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", config.EMAIL_SENDER)
	m.SetHeader("To", recipientEmail)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body.String())

	d := gomail.NewDialer("smtp.gmail.com", 587, config.EMAIL_SENDER, initializers.CONFIG.GMAIL_KEY)

	if err := d.DialAndSend(m); err != nil {
		LogDatabaseError("Error while sending Mail", err, "go_routine")
	}
}
