package helpers

import (
	"fmt"
	"regexp"
	"unicode"

	goaway "github.com/TwiN/go-away"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
)

var Profanities = []string{
	"fag",
	"fuck",
	"nigga",
	"nigger",
	"rape",
	"retard",
}

var profanityDetector = goaway.NewProfanityDetector().WithCustomDictionary(Profanities, []string{}, []string{})

// FirstDigitRegexp returns the first digit found in a string (used for checking duplicate slugs)
var FirstDigitRegexp = regexp.MustCompile(`\d`)

var emailRegexp = regexp.MustCompile("^.+?@.+?...+?$")

var PasswordRegExp = regexp.MustCompile("[`!@#$%^&*()_+=\\[\\]{};':\"\\|,.<>\\/?~]")

var UsernameRegExp = regexp.MustCompile("[`~!@#$%^&*()+=\\[\\]{};':\"\\|,.<>\\/?]")

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
	if PasswordRegExp.Match([]byte(password)) {
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

	err := fmt.Errorf("password does not meet requirements")
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

// PrintableLen counts how many printable characters are in a string.
func PrintableLen(s string) int {
	sLen := 0
	runes := []rune(s)
	for _, r := range runes {
		if unicode.IsPrint(r) {
			sLen += 1
		}
	}
	return sLen
}

func IsValidUsername(username string) error {
	// Must contain at least 3 characters
	// Cannot contain more than 15 characters
	// Cannot contain profanity
	// Can only contain the following symbols: _
	hasDisallowedSymbol := false
	if UsernameRegExp.Match([]byte(username)) {
		hasDisallowedSymbol = true
	}

	err := fmt.Errorf("username does not meet requirements")
	if TrimUsername(username) == "" {
		return terror.Error(err, "Invalid username. Your username cannot be empty.")
	}
	if PrintableLen(TrimUsername(username)) < 3 {
		return terror.Error(err, "Invalid username. Your username must be at least 3 characters long.")
	}
	if PrintableLen(TrimUsername(username)) > 30 {
		return terror.Error(err, "Invalid username. Your username cannot be more than 30 characters long.")
	}
	if hasDisallowedSymbol {
		return terror.Error(err, "Invalid username. Your username contains a disallowed symbol.")
	}
	if goaway.IsProfane(username) {
		return terror.Error(err, "Invalid username. Your username contains profanity.")
	}

	return nil
}
