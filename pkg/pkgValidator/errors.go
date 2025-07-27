package pkgValidator

import "errors"

var (
	ErrInvalidInput             = errors.New("invalid input")
	ErrInvalidEmail             = errors.New("invalid email")
	ErrPasswordTooWeak          = errors.New("password must be at least 8 characters")
	ErrInvalidRole              = errors.New("role must be employee or moderator")
	ErrInvalidCity              = errors.New("city must be Moscow, Saint Petersburg or Kazan")
	ErrUserExists               = errors.New("user already exists")
	ErrInvalidCredentials       = errors.New("invalid credentials")
	ErrInvalidReceptionCreation = errors.New("pvz's last reception is still in progress")
	ErrInvalidPVZID             = errors.New("invalid pvz_id")
	ErrInvalidProductType       = errors.New("type must be electronics, clothes or shoes")
	ErrNoActiveReception        = errors.New("no active reception found")
	ErrInvalidPage              = errors.New("page must be greater than 0")
	ErrInvalidLimit             = errors.New("limit must be between 1 and 30")
	ErrInvalidDateFormat        = errors.New("date must be in RFC3339 format")
	ErrInvalidDateRange         = errors.New("end date must be after start date")
	ErrLimitTooHigh = errors.New("limit cannot be higher than 30")
)
