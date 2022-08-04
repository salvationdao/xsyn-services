package email

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
	"xsyn-services/types"

	"github.com/aymerick/raymond"
	"github.com/mailgun/mailgun-go/v4"
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
	PassportWebHostURL string

	// Handlebars Email Templates
	Templates map[string]*raymond.Template

	// Path to template
	TemplatePath string
}

// NewMailer returns a new Mailer controller
func NewMailer(domain string, apiKey string, systemAddress string, config *types.Config, log *zerolog.Logger) (*Mailer, error) {
	mailer := &Mailer{
		MailGun:            mailgun.NewMailgun(domain, apiKey),
		Log:                log,
		SystemAddress:      systemAddress,
		PassportWebHostURL: config.PassportWebHostURL,
		Templates:          map[string]*raymond.Template{},
		TemplatePath:       config.EmailTemplatePath,
	}

	// Handlebar template helpers
	raymond.RegisterHelper("logo", func() raymond.SafeString {
		return raymond.SafeString(`<img src="cid:logo.png" alt="Passport" height="40px" />`)
	})

	// Parse email templates
	var templates []string
	err := filepath.Walk(config.EmailTemplatePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return terror.Error(err, "failed to walk through templates folder")
		}
		if strings.Contains(path, ".html") {
			templates = append(templates, path)
		}
		return nil
	})
	if err != nil {
		return nil, terror.Error(err, "failed to step through templates folder")
	}
	if len(templates) == 0 {
		return nil, terror.Error(fmt.Errorf("unable to find templates"), "zero templates found")

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
	ctx context.Context,
	to string,
	subject string,
	template string,
	content interface{},
	bcc string,
	attachments ...types.Blob,
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
	message.AddInline(fmt.Sprintf("%s/logo.png", m.TemplatePath))
	if bcc != "" {
		for _, b := range strings.Split(bcc, ",") {
			message.AddBCC(b)
		}
	}
	for _, attachment := range attachments {
		message.AddBufferAttachment(attachment.FileName+"."+attachment.Extension, attachment.File)
	}

	// Send Email
	emailCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	_, _, err = m.MailGun.Send(emailCtx, message)
	if err != nil {
		return terror.Error(err, "failed to send email")
	}

	return nil
}
