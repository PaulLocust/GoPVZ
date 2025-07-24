package pkgValidator

import "errors"

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrPasswordTooWeak    = errors.New("password must be at least 8 characters")
	ErrInvalidRole        = errors.New("role must be employee or moderator")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
