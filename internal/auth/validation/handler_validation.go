package validation

import (
	"strings"

	"GoPVZ/internal/dto"
	"GoPVZ/pkg/pkgValidator"
)

type DummyLoginValidator struct {
	Payload dto.PostDummyLoginJSONBody
}

func NewDummyLoginValidator(payload dto.PostDummyLoginJSONBody) *DummyLoginValidator {
	return &DummyLoginValidator{Payload: payload}
}

func (v *DummyLoginValidator) Validate() error {
	if v.Payload.Role != dto.PostDummyLoginJSONBodyRoleEmployee && v.Payload.Role != dto.PostDummyLoginJSONBodyRoleModerator {
		return pkgValidator.ErrInvalidRole
	}

	return nil
}

type RegisterValidator struct {
	Payload dto.PostRegisterJSONBody
}

func NewRegisterValidator(payload dto.PostRegisterJSONBody) *RegisterValidator {
	return &RegisterValidator{Payload: payload}
}

func (v *RegisterValidator) Validate() error {

	if v.Payload.Email == "" || !strings.Contains(string(v.Payload.Email), "@") {
		return pkgValidator.ErrInvalidEmail
	}
	if len(v.Payload.Password) < 8 {
		return pkgValidator.ErrPasswordTooWeak
	}
	if v.Payload.Role != dto.Employee && v.Payload.Role != dto.Moderator {
		return pkgValidator.ErrInvalidRole
	}

	return nil
}

type LoginValidator struct {
	Payload dto.PostLoginJSONBody
}

func NewLoginValidator(payload dto.PostLoginJSONBody) *LoginValidator {
	return &LoginValidator{Payload: payload}
}

func (v *LoginValidator) Validate() error {

	if v.Payload.Email == "" || !strings.Contains(string(v.Payload.Email), "@") {
		return pkgValidator.ErrInvalidEmail
	}
	if len(v.Payload.Password) < 8 {
		return pkgValidator.ErrPasswordTooWeak
	}

	return nil
}
