package email

const (
	provGoogle = "google"
	provAWS    = "aws"
)

// Sender is the object to use for Email
type Sender interface {
	Send(*Email) *Error
}

// NewSender creates a new sender depending on the provider specified in EmailConfig
func NewSender() (Sender, *Error) {
	if conf.Provider == provGoogle {
		return NewGmail(conf)
	}
	return nil, NewError("Provider Not Implemented", nil)
}
