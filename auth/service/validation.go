package service

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

var defaultValidator *Validator

func init() {
	defaultValidator = NewValidator()
}

func Default() *Validator {
	return defaultValidator
}

func NewValidator() *Validator {
	v := validator.New()

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{validate: v}
}

func (v *Validator) Struct(s interface{}) *Report {
	err := v.validate.Struct(s)
	if err == nil {
		return &Report{}
	}

	verrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return &Report{
			Errors: []FieldError{
				{Field: "unknown", Actual: "internal", Detail: err.Error()},
			},
		}
	}

	return convertErrors(verrs)
}

func (v *Validator) Var(field interface{}, tag string) *Report {
	err := v.validate.Var(field, tag)
	if err == nil {
		return &Report{}
	}

	verrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return &Report{
			Errors: []FieldError{
				{Field: "unknown", Actual: "internal", Detail: err.Error()},
			},
		}
	}

	return convertErrors(verrs)
}

func (v *Validator) RegisterValidation(tag string, fn validator.Func) error {
	return v.validate.RegisterValidation(tag, fn)
}

func (v *Validator) RegisterStructLevelValidation(fn validator.StructLevelFunc, types ...interface{}) {
	v.validate.RegisterStructValidation(fn, types...)
}

func (v *Validator) Engine() *validator.Validate {
	return v.validate
}

// ────────────────────────── Report / Converter ──────────────────────────

type FieldError struct {
	Field  string `json:"field"`
	Actual string `json:"actual"`
	Detail string `json:"detail"`
}

type Report struct {
	Errors []FieldError `json:"errors"`
}

func (r *Report) HasErrors() bool {
	return len(r.Errors) > 0
}

func (r *Report) Error() string {
	if !r.HasErrors() {
		return ""
	}
	var sb strings.Builder
	for i, fe := range r.Errors {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(fe.Field)
		sb.WriteString(": ")
		sb.WriteString(fe.Detail)
	}
	return sb.String()
}

func (r *Report) ToMap() map[string]string {
	m := make(map[string]string, len(r.Errors))
	for _, fe := range r.Errors {
		m[fe.Field] = fe.Detail
	}
	return m
}

func (r *Report) IsEmpty() bool {
	return !r.HasErrors()
}

// ─────────────────────── internal helpers ───────────────────────

func convertErrors(verrs validator.ValidationErrors) *Report {
	errors := make([]FieldError, 0, len(verrs))
	for _, fe := range verrs {
		errors = append(errors, FieldError{
			Field:  fe.Field(),
			Actual: fe.ActualTag(),
			Detail: buildDetail(fe),
		})
	}
	return &Report{Errors: errors}
}

func buildDetail(fe validator.FieldError) string {
	param := fe.Param()
	tag := fe.ActualTag()

	switch tag {
	case "required":
		return "field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		if param != "" {
			return fmt.Sprintf("must be at least %s", param)
		}
		return "value is below minimum"
	case "max":
		if param != "" {
			return fmt.Sprintf("must be at most %s", param)
		}
		return "value is above maximum"
	case "len":
		return fmt.Sprintf("must be exactly %s characters long", param)
	case "eq":
		return fmt.Sprintf("must equal %s", param)
	case "ne":
		return fmt.Sprintf("must not equal %s", param)
	case "lt":
		return fmt.Sprintf("must be less than %s", param)
	case "lte":
		return fmt.Sprintf("must be less than or equal to %s", param)
	case "gt":
		return fmt.Sprintf("must be greater than %s", param)
	case "gte":
		return fmt.Sprintf("must be greater than or equal to %s", param)
	case "eqfield":
		return fmt.Sprintf("must equal %s", param)
	case "nefield":
		return fmt.Sprintf("must not equal %s", param)
	case "oneof":
		return fmt.Sprintf("must be one of [%s]", param)
	case "url":
		return "must be a valid URL"
	case "uuid":
		return "must be a valid UUID"
	case "numeric":
		return "must be numeric"
	case "alpha":
		return "must contain only letters"
	case "alphanum":
		return "must contain only letters and numbers"
	case "boolean":
		return "must be a boolean value"
	case "json":
		return "must be valid JSON"
	default:
		if param != "" {
			return fmt.Sprintf("failed validation on %s (%s)", tag, param)
		}
		return fmt.Sprintf("failed validation on %s", tag)
	}
}
