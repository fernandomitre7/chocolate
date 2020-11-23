package email

import (
	"fmt"
	"net/smtp"

	"chocolate/service/shared/logger"

	"chocolate/service/shared/config"
)

// Gmail is the provider to send emails thru google mail
type Gmail struct {
	Config config.EmailConfig
}

// NewGmail creates a new Gmail sender
func NewGmail(config config.EmailConfig) (*Gmail, *Error) {
	if config.Auth.Username == "" || config.Auth.Password == "" ||
		config.Host == "" || config.Port == "" {
		return nil, NewError("Missing Configuration fields", nil)
	}
	return &Gmail{
		Config: config,
	}, nil
}

// Send sends an email
func (g Gmail) Send(email *Email) (eErr *Error) {

	//mime := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n"

	//msg := []byte(subject + mime + "\n" + email.Body)
	msg, err := email.FormMessage()
	if err != nil {
		logger.Errorf("Couldn't form email body message: %s", err.Error())
		eErr = NewError("Couldn't form email body message", err)
		return
	}
	logger.Debugf("Email Message: %s", msg)
	// Set up authentication information.
	auth := smtp.PlainAuth(
		"",
		g.Config.Auth.Username,
		g.Config.Auth.Password,
		g.Config.Host,
	)
	addr := fmt.Sprintf("%s:%s", g.Config.Host, g.Config.Port)
	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	err = smtp.SendMail(
		addr,
		auth,
		email.From,
		append(email.CC, email.To),
		[]byte(msg),
	)
	if err != nil {
		logger.Errorf("Couldn't send Email: %+v", err)
		eErr = NewError("Couldn't send Email", err)
	}
	return

}
