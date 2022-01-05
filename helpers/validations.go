package helpers

import (
	"fmt"
	"regexp"

	gocheckpasswd "github.com/ninja-software/go-check-passwd"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
)

// FirstDigitRegexp returns the first digit found in a string (used for checking duplicate slugs)
var FirstDigitRegexp = regexp.MustCompile(`\d`)

var emailRegexp = regexp.MustCompile("^.+?@.+?\\..+?$")

// IsEmpty checks if string given is empty
func IsEmpty(text *null.String) bool {
	return text == nil || text.String == ""
}

// IsEmptyStringPtr checks if string pointer is empty
func IsEmptyStringPtr(text *string) bool {
	return text == nil || *text == ""
}

// IsValidEmail checks if email given is valid
func IsValidEmail(email string) bool {
	return emailRegexp.MatchString(email)
}

// IsValidPassword checks whether the password entered is valid
func IsValidPassword(password string) error {
	if len(password) < 8 {
		return terror.Error(fmt.Errorf("password must contain at least 8 characters"), "Passwords must contain at least 8 characters")
	}
	if gocheckpasswd.IsCommon(password) {
		return terror.Error(fmt.Errorf("password is common, try another one"), "Passwords entered is commonly used, please try another")
	}
	return nil
}
