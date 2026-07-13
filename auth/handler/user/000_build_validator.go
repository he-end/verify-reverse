package user

import (
	"regexp"

	"github.com/go-playground/validator/v10"

	"github.com/he-end/verify-reverse/auth/service"
)

var phonePattern = regexp.MustCompile(`^\+?[\d\s\-]{7,20}$`)

func (h *Handler) registerValidator(val *service.Validator) {
	val.RegisterStructLevelValidation(regValUpdateProfile, UpdateProfileReqBody{})
	val.RegisterStructLevelValidation(regValChangePassword, ChangePasswordReqBody{})
	val.RegisterStructLevelValidation(regValChangeWANumber, ChangeWANumberReqBody{})
}

func regValUpdateProfile(sl validator.StructLevel) {
	req := sl.Current().Interface().(UpdateProfileReqBody)
	if req.Name == "" {
		sl.ReportError(req.Name, "name", "Name", "required", "")
	} else {
		if len(req.Name) < 2 {
			sl.ReportError(req.Name, "name", "Name", "min", "2")
		}
		if len(req.Name) > 100 {
			sl.ReportError(req.Name, "name", "Name", "max", "100")
		}
	}
}

func regValChangePassword(sl validator.StructLevel) {
	req := sl.Current().Interface().(ChangePasswordReqBody)
	if req.OldPassword == "" {
		sl.ReportError(req.OldPassword, "old_password", "OldPassword", "required", "")
	}
	if req.NewPassword == "" {
		sl.ReportError(req.NewPassword, "new_password", "NewPassword", "required", "")
	} else {
		if len(req.NewPassword) < 8 {
			sl.ReportError(req.NewPassword, "new_password", "NewPassword", "min", "8")
		}
		if len(req.NewPassword) > 128 {
			sl.ReportError(req.NewPassword, "new_password", "NewPassword", "max", "128")
		}
		if !passwordRequired(req.NewPassword) {
			sl.ReportError(req.NewPassword, "new_password", "NewPassword", "passwordcomplex", "")
		}
	}
	if req.ConfirmPassword == "" {
		sl.ReportError(req.ConfirmPassword, "confirm_password", "ConfirmPassword", "required", "")
	} else if req.NewPassword != req.ConfirmPassword {
		sl.ReportError(req.ConfirmPassword, "confirm_password", "ConfirmPassword", "eqfield", "new_password")
	}
}

func regValChangeWANumber(sl validator.StructLevel) {
	req := sl.Current().Interface().(ChangeWANumberReqBody)
	if req.Number == "" {
		sl.ReportError(req.Number, "number", "Number", "required", "")
	} else if !phonePattern.MatchString(req.Number) {
		sl.ReportError(req.Number, "number", "Number", "phone", "")
	}
}

func passwordRequired(pwd string) bool {
	var hasUpper, hasDigit, hasSpecial bool
	for _, r := range pwd {
		switch {
		case r >= 'A' && r <= 'Z':
			hasUpper = true
		case r >= '0' && r <= '9':
			hasDigit = true
		case (r >= 33 && r <= 47) || (r >= 58 && r <= 64) || (r >= 91 && r <= 96) || (r >= 123 && r <= 126):
			hasSpecial = true
		}
	}
	return hasUpper && hasDigit && hasSpecial
}
