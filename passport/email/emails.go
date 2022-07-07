package email

import (
	"context"
	"fmt"
	"xsyn-services/types"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
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
func (m *Mailer) SendForgotPasswordEmail(ctx context.Context, user *types.User, token string, tokenID uuid.UUID) error {
	hostURL := m.PassportWebHostURL

	err := m.SendEmail(ctx,
		user.Email.String,
		"Forgot Password - Passport XSYN",
		"forgot_password",
		struct {
			MagicLink string `handlebars:"magic_link"`
			Name      string `handlebars:"name"`
		}{
			MagicLink: fmt.Sprintf("%s/reset-password?id=%s&token=%s", hostURL, tokenID, token),
			Name:      user.Username,
		},
		"",
	)
	if err != nil {
		return terror.Error(err, " Failed to send forgot password email")
	}
	return nil
}

// SendVerificationEmail sends an email with the confirm_email template
func (m *Mailer) SendVerificationEmail(ctx context.Context, user *types.User, token string, newAccount bool) error {
	hostURL := m.PassportWebHostURL

	err := m.SendEmail(ctx,
		user.Email.String,
		"Verify Email",
		"confirm_email",
		struct {
			MagicLink  string      `handlebars:"magic_link"`
			Name       string      `handlebars:"name"`
			Email      null.String `handlebars:"email"`
			NewAccount bool        `handlebars:"new_account"`
		}{
			MagicLink:  fmt.Sprintf("%s/verify?token=%s", hostURL, token),
			Name:       user.Username,
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
