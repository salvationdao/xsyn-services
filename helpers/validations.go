package helpers

import (
	"fmt"
	"regexp"
	"unicode"

	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
)

// FirstDigitRegexp returns the first digit found in a string (used for checking duplicate slugs)
var FirstDigitRegexp = regexp.MustCompile(`\d`)

var emailRegexp = regexp.MustCompile("^.+?@.+?...+?$")

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
	// Must contain at least 8 characters
	// Must contain at least 1 upper and 1 lower case letter
	// Must contain at least 1 number
	// Must contain at least one symbol
	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSymbol := false
	reg, err := regexp.Compile("[`!@#$%^&*()_+\\-=\\[\\]{};':\"\\|,.<>\\/?~]")
	if err != nil {
		return terror.Error(err, "Something went wrong. Please try again.")
	}
	if reg.Match([]byte(password)) {
		hasSymbol = true
	}
	for _, r := range password {
		if unicode.IsUpper(r) {
			hasUpper = true
		} else if unicode.IsLower(r) {
			hasLower = true
		} else if unicode.IsNumber(r) {
			hasNumber = true
		}
	}

	err = fmt.Errorf("password does not meet requirements")
	if len(password) < 8 {
		return terror.Error(err, "Invalid password. Your password must be at least 8 characters long.")
	}
	if !hasNumber {
		return terror.Error(err, "Invalid password. Your password must contain at least 1 number.")
	}
	if !hasUpper {
		return terror.Error(err, "Invalid password. Your password must contain at least 1 uppercase letter.")
	}
	if !hasLower {
		return terror.Error(err, "Invalid password. Your password must contain at least 1 lowercase letter.")
	}
	if !hasSymbol {
		return terror.Error(err, "Invalid password. Your password must contain at least 1 symbol.")
	}
	return nil
}
