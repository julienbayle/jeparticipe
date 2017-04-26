package email

import (
	"bytes"
	"html/template"
)

// Email struct
type Email struct {
	To      string
	Subject string
	Body    string
}

// Creates a new email
func NewEmail(to string, subject, body string) *Email {
	return &Email{
		To:      to,
		Subject: subject,
		Body:    body,
	}
}

// Updates body using a HTML template
func (email *Email) AddBodyUsingTemplate(templateFileName string, data interface{}) {
	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		panic("Missing template " + templateFileName)
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		panic("Template building process failed " + templateFileName)
	}
	email.Body = buf.String()
}
