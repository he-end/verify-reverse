package auth

import (
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/he-end/verify-reverse/verify/service"
)

var phonePattern = regexp.MustCompile(`^\+?[\d\s\-]{7,20}$`)
var emailPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func (h *Handler) registerValidator(val *service.Validator) {
	val.RegisterStructLevelValidation(regValRegisterViaWA, RegisterViaWAReqBody{})
	val.RegisterStructLevelValidation(regValRegisterViaEmail, RegisterViaEmailReqBody{})
	val.RegisterStructLevelValidation(regValLogin, LoginReqBody{})
}

func regValRegisterViaWA(sl validator.StructLevel) {
	reg := sl.Current().Interface().(RegisterViaWAReqBody)
	if reg.Name == nil {
		sl.ReportError(reg.Name, "name", "Name", "required", "")
	} else {
		if len(*reg.Name) > 100 {
			sl.ReportError(reg.Name, "name", "Name", "max", "100")
		}
		if len(*reg.Name) < 2 {
			sl.ReportError(reg.Name, "name", "Name", "min", "2")
		}
	}
	if reg.Number == nil {
		sl.ReportError(reg.Number, "number", "Number", "required", "")
	} else if !phonePattern.MatchString(*reg.Number) {
		sl.ReportError(reg.Number, "number", "Number", "phone", "")
	}
	if reg.Pwd != nil {
		if len(*reg.Pwd) < 8 {
			sl.ReportError(reg.Pwd, "password", "Pwd", "min", "8")
		}
		if len(*reg.Pwd) > 128 {
			sl.ReportError(reg.Pwd, "password", "Pwd", "max", "128")
		}
		if !passwordRequired(*reg.Pwd) {
			sl.ReportError(reg.Pwd, "password", "Pwd", "passwordcomplex", "")
		}
	}
	if reg.Pwd != nil && reg.ConfirmPwd != nil && *reg.ConfirmPwd != *reg.Pwd {
		sl.ReportError(reg.ConfirmPwd, "confirm_password", "ConfirmPwd", "eqfield", "password")
	}
}

func regValRegisterViaEmail(sl validator.StructLevel) {
	reg := sl.Current().Interface().(RegisterViaEmailReqBody)
	if reg.Email == nil {
		sl.ReportError(reg.Email, "email", "Email", "required", "")
	} else if !emailPattern.MatchString(*reg.Email) {
		sl.ReportError(reg.Email, "email", "Email", "email", "")
	}
	if reg.Name == nil {
		sl.ReportError(reg.Name, "name", "Name", "required", "")
	} else {
		if len(*reg.Name) > 100 {
			sl.ReportError(reg.Name, "name", "Name", "max", "100")
		}
		if len(*reg.Name) < 2 {
			sl.ReportError(reg.Name, "name", "Name", "min", "2")
		}
	}
	if reg.Pwd == nil {
		sl.ReportError(reg.Pwd, "password", "Pwd", "required", "")
	} else {
		if len(*reg.Pwd) < 8 {
			sl.ReportError(reg.Pwd, "password", "Pwd", "min", "8")
		}
		if len(*reg.Pwd) > 128 {
			sl.ReportError(reg.Pwd, "password", "Pwd", "max", "128")
		}
		if !passwordRequired(*reg.Pwd) {
			sl.ReportError(reg.Pwd, "password", "Pwd", "passwordcomplex", "")
		}
	}
	if reg.ConfirmPwd == nil {
		sl.ReportError(reg.ConfirmPwd, "confirm_password", "ConfirmPwd", "required", "")
	} else if reg.Pwd != nil && *reg.ConfirmPwd != *reg.Pwd {
		sl.ReportError(reg.ConfirmPwd, "confirm_password", "ConfirmPwd", "eqfield", "password")
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

func regValLogin(sl validator.StructLevel) {
	reg := sl.Current().Interface().(LoginReqBody)
	if reg.Email == nil && reg.Number == nil {
		sl.ReportError(reg.Email, "email", "Email", "required_with", "number")
		sl.ReportError(reg.Number, "number", "Number", "required_with", "email")
	}
	if reg.Pwd == "" {
		sl.ReportError(reg.Pwd, "password", "Pwd", "required", "")
	}
}
