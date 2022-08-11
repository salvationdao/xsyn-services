package email

import (
	"context"
	"fmt"
	"strings"
	"xsyn-services/passport/passlog"
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
		passlog.L.Error().Err(err).Msg("failed to send basic email")
		return terror.Error(err, "Failed to send basic email")
	}
	return nil
}

// SendForgotPasswordEmail sends an email with the forgot_password template
func (m *Mailer) SendForgotPasswordEmail(ctx context.Context, user *types.User, token string) error {
	hostURL := m.PassportWebHostURL

	err := m.SendEmail(ctx,
		user.Email.String,
		"Forgot Password - Passport XSYN",
		"forgot_password",
		struct {
			MagicLink string `handlebars:"magic_link"`
			Name      string `handlebars:"name"`
		}{
			MagicLink: fmt.Sprintf("%s/reset-password?token=%s", hostURL, token),
			Name:      user.Username,
		},
		"",
	)
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to send forgot password email")
		return terror.Error(err, " Failed to send forgot password email")
	}
	return nil
}

// SendVerificationEmail sends an email with the confirm_email template
func (m *Mailer) SendVerificationEmail(ctx context.Context, user *types.User, code string) error {

	err := m.SendEmail(ctx,
		user.Email.String,
		"Verify Email  - Passport XSYN",
		"confirm_email",
		struct {
			Code  string `handlebars:"code"`
			Name  string `handlebars:"name"`
			Email string `handlebars:"email"`
		}{
			Code:  strings.ToUpper(code),
			Name:  user.Username,
			Email: user.Email.String,
		},
		"",
	)
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to send verify email")
		return terror.Error(err, "Failed to send verification email")
	}
	return nil
}

// SendSignupEmail sends an email with the signup template
func (m *Mailer) SendSignupEmail(ctx context.Context, email string, code string) error {
	err := m.SendEmail(ctx,
		email,
		"New user please verify email  - Passport XSYN",
		"signup",
		struct {
			Code  string `handlebars:"code"`
			Email string `handlebars:"email"`
		}{
			Code:  strings.ToUpper(code),
			Email: email,
		},
		"",
	)
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to send signup email")
		return terror.Error(err, "Failed to send signup email")
	}
	return nil
}
