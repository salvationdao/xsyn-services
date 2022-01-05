package email

import (
	"passport"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aymerick/raymond"
	"github.com/mailgun/mailgun-go/v3"
	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"
)

// Mailer controller
type Mailer struct {
	MailGun *mailgun.MailgunImpl
	Log     *zerolog.Logger

	// Email address that the system sends from
	SystemAddress string
	// AdminHostURL used for links
	AdminHostURL string
	// PublicHostURL used for links
	PublicHostURL string

	// Handlebars Email Templates
	Templates map[string]*raymond.Template
}

// NewMailer returns a new Mailer controller
func NewMailer(domain string, apiKey string, systemAddress string, config *passport.Config, log *zerolog.Logger) (*Mailer, error) {
	mailer := &Mailer{
		MailGun:       mailgun.NewMailgun(domain, apiKey),
		Log:           log,
		SystemAddress: systemAddress,
		AdminHostURL:  config.AdminHostURL,
		PublicHostURL: config.PublicHostURL,
		Templates:     map[string]*raymond.Template{},
	}

	// Handlebar template helpers
	raymond.RegisterHelper("logo", func() raymond.SafeString {
		return raymond.SafeString(`<img src="cid:logo.png" alt="Passport" height="40px" />`)
	})

	// Parse email templates
	var templates []string
	templatesFolder := "./email/templates"
	err := filepath.Walk(templatesFolder, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".html") {
			templates = append(templates, path)
		}
		return nil
	})
	if err != nil {
		return nil, terror.Error(err, "failed to step through templates folder")
	}
	err = mailer.ParseTemplates(templates...)
	if err != nil {
		return nil, terror.Error(err, "failed to parse email templates")
	}
	log.Info().Msg("Mailgun initialized.")
	return mailer, nil
}

// ParseTemplates parses email template files and adds them to the Mailer's Templates map
func (m *Mailer) ParseTemplates(fileNames ...string) error {
	for _, fileName := range fileNames {
		template, err := ParseTemplate(fileName)
		if err != nil {
			return terror.Error(err, "")
		}
		slashIndex := strings.LastIndex(fileName, "/")
		if slashIndex == -1 {
			slashIndex = strings.LastIndex(fileName, "\\")
		}
		name := fileName[slashIndex+1 : strings.LastIndex(fileName, ".")]
		m.Templates[name] = template
	}
	return nil
}

// ParseTemplate reads a template file and parses it ready for rendering
func ParseTemplate(fileName string) (*raymond.Template, error) {
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, terror.Error(err, "failed to read template file: "+fileName)
	}
	template, err := raymond.Parse(string(file))
	if err != nil {
		return nil, terror.Error(err, "failed to parse template: "+fileName)
	}
	return template, nil
}

// SendEmail sends a templated email via smtp. attachments must be the full file path
func (m *Mailer) SendEmail(
	to string,
	subject string,
	template string,
	content interface{},
	bcc string,
	attachments ...passport.Blob,
) error {
	// Get Template
	emailTemplate, ok := m.Templates[template]
	if !ok {
		return terror.Error(terror.ErrInvalidInput, "invalid email template name")
	}

	// Render template
	body, err := emailTemplate.Exec(content)
	if err != nil {
		return terror.Error(err, "failed to render template")
	}

	// Setup Email
	message := m.MailGun.NewMessage(m.SystemAddress, subject, "", strings.Split(to, ",")...)
	message.SetHtml(body)
	message.AddInline("./email/templates/logo.png")
	if bcc != "" {
		for _, b := range strings.Split(bcc, ",") {
			message.AddBCC(b)
		}
	}
	for _, attachment := range attachments {
		message.AddBufferAttachment(attachment.FileName+"."+attachment.Extension, attachment.File)
	}

	// Send Email
	emailCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	_, _, err = m.MailGun.Send(emailCtx, message)
	if err != nil {
		return terror.Error(err, "failed to send email")
	}

	return nil
}
