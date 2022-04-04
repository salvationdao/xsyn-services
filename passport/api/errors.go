package api

import "fmt"

// ErrNotImplemented used as placeholder
var ErrNotImplemented = fmt.Errorf("not implemented")

// ErrClientHasNoUser is returned when the elevate receives a nil user
var ErrClientHasNoUser = fmt.Errorf("websocket client has no user")

// ErrTokenBlacklisted is returned when an issue token is not found
var ErrTokenBlacklisted = fmt.Errorf("token is blacklisted")

// ErrUserAlreadyVerified is returned when processing an unverified user that is already verified
var ErrUserAlreadyVerified = fmt.Errorf("user is already verified")
