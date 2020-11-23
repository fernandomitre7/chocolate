package email

import (
	"fmt"
	"strings"

	"chocolate/service/shared/config"
)

var conf config.EmailConfig

var Templates map[string]string

// Init initalizes package with correct configurations
func Init(cnf config.EmailConfig) {
	conf = cnf
	Templates = cnf.Templates
}

// Type of the type of email we are going to send
type Type string

const (
	// TextEmail type for plain text emails
	TextEmail = Type("text")
	// HTMLEmail type for HTML emails
	HTMLEmail = Type("html")
)

// Email is the email object to send
type Email struct {
	Type     Type
	Subject  string
	Body     string
	From     string
	To       string
	CC       []string
	Template ITemplate
}

// ITemplate is the interface to use to be able to have diffent types of templates for emails
type ITemplate interface {
	Process() (string, error)
}

// FormMessage properly forms the Email message in RFC 822-style
func (e Email) FormMessage() (msg string, err error) {
	/*
		From: example@example.com
		To: example2@example.com
		Subject: As basic as it gets

		This is the plain text body of the message.  Note the blank line
		between the header information and the body of the message.
	*/
	var contentType, header, body string
	if e.Type == TextEmail {
		body = e.Body
		// "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n"
		contentType = `text/plain; charset="UTF-8"`
	} else {
		if body, err = e.Template.Process(); err != nil {
			return "", err
		}
		contentType = `text/html; charset="UTF-8"`
	}
	header = fmt.Sprintf("MIME-Version: 1.0;\nContent-Type:%s;\n", contentType)
	if e.CC != nil && len(e.CC) > 0 {
		cc := strings.Join(e.CC, ", ")
		msg = fmt.Sprintf("%sFrom: %s\nTo: %s\nCc: %s\nSubject: %s\n\n%s", header, e.From, e.To, cc, e.Subject, body)
	} else {
		msg = fmt.Sprintf("%sFrom: %s\nTo: %s\nSubject: %s\n\n%s", header, e.From, e.To, e.Subject, body)
	}
	return
}
