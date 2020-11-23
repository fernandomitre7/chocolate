package confirm

import (
	"bytes"
	"html/template"

	"chocolate/service/shared/email"
)

// Template is the template for confirmaiton emails
type Template struct {
	Location string
	Data     TemplateData
}

// TemplateData is the data structure for confirmation email
type TemplateData struct {
	Username   string
	ConfirmURL string
}

// NewTemplate creates a confirmation template
func NewTemplate(username, confirmURL string) *Template {
	return &Template{
		Location: email.Templates["confirm"],
		Data: TemplateData{
			Username:   username,
			ConfirmURL: confirmURL,
		},
	}
}

// Process returns the string ot the template with the data
func (ct Template) Process() (string, error) {
	t, err := template.ParseFiles(ct.Location)
	if err != nil {
		return "", err
	}
	/* t := template.New("confirm")      //name of the template is main
	t, err = t.Parse(confirmTemplate) // parsing of template string */
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, ct.Data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
