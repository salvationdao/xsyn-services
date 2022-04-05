package email

import (
	"context"
	"fmt"
	"xsyn-services/types"

	"github.com/ninja-software/terror/v2"
)

type HostFrom = string

const (
	HostFromPublic HostFrom = "PUBLIC"
	HostFromAdmin  HostFrom = "ADMIN"
)

type User struct {
	IsAdmin   bool
	Email     string
	FirstName string
	LastName  string
}

// SendBasicEmail sends a plain text email with the basic template
func (m *Mailer) SendBasicEmail(ctx context.Context, to string, subject string, message string, attachments ...types.Blob) error {
	err := m.SendEmail(
		ctx,
		to,
		subject,
		"basic",
		struct {
			Message string `handlebars:"message"`
		}{
			Message: message,
		},
		"",
		attachments...,
	)
	if err != nil {
		return terror.Error(err, "Failed to send basic email")
	}
	return nil
}

// SendForgotPasswordEmail sends an email with the forgot_password template
func (m *Mailer) SendForgotPasswordEmail(ctx context.Context, user *User, token string) error {
	hostURL := m.PassportWebHostURL
	if user.IsAdmin {
		hostURL = m.PassportWebHostURL
	}
	err := m.SendEmail(ctx,
		user.Email,
		"Forgot Password",
		"forgot_password",
		struct {
			MagicLink string `handlebars:"magic_link"`
			Name      string `handlebars:"name"`
		}{
			MagicLink: fmt.Sprintf("%s/verify?token=%s&forgot=true", hostURL, token),
			Name:      user.FirstName + " " + user.LastName,
		},
		"",
	)
	if err != nil {
		return terror.Error(err, " Failed tosend forgot password email")
	}
	return nil
}

// SendVerificationEmail sends an email with the confirm_email template
func (m *Mailer) SendVerificationEmail(ctx context.Context, user *User, token string, newAccount bool) error {
	hostURL := m.PassportWebHostURL
	if user.IsAdmin {
		hostURL = m.PassportWebHostURL
	}
	err := m.SendEmail(ctx,
		user.Email,
		"Verify Email",
		"confirm_email",
		struct {
			MagicLink  string `handlebars:"magic_link"`
			Name       string `handlebars:"name"`
			Email      string `handlebars:"email"`
			NewAccount bool   `handlebars:"new_account"`
		}{
			MagicLink:  fmt.Sprintf("%s/verify?token=%s", hostURL, token),
			Name:       user.FirstName + " " + user.LastName,
			Email:      user.Email,
			NewAccount: newAccount,
		},
		"",
	)
	if err != nil {
		return terror.Error(err, "Failed to send verification email")
	}
	return nil
}
