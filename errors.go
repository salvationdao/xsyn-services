package passport

import (
	"fmt"
)

// ErrPasswordShort when the entered password is too short
var ErrPasswordShort = fmt.Errorf("password too short")

// ErrPasswordCommon when the entered password is too common
var ErrPasswordCommon = fmt.Errorf("password too common")

// ErrEmailInvalid when the entered email is invalid
var ErrEmailInvalid = fmt.Errorf("invalid email")

// ErrDataReserved when the targeted data is reserved
var ErrDataReserved = fmt.Errorf("data is reserved")

// ErrEmailUsed when trying to register with an email that's in use
var ErrEmailUsed = fmt.Errorf("email used")
