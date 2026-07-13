package repository

import "errors"

var (
	ErrNotFound     = errors.New("record not found")
	ErrDuplicateKey = errors.New("duplicate key violates unique constraint")
)

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrMissingContact      = errors.New("email or phone number is required")
	ErrTokenExpired        = errors.New("token has expired")
	ErrTokenInvalid        = errors.New("token is invalid")
	ErrVerificationExpired = errors.New("verification code has expired")
	ErrVerificationPending = errors.New("verification already pending for this contact")
	ErrVerificationUsed    = errors.New("verification code already used")
)

var (
	ErrTooManyAttempts      = errors.New("too many verification attempts")
	ErrSenderBlocked        = errors.New("sender is temporarily blocked")
	ErrAlreadyVerified      = errors.New("contact already verified")
	ErrVerificationNotValid = errors.New("verification not valid")
)

var (
	ErrWrongPassword = errors.New("wrong password")
	ErrNumberTaken   = errors.New("phone number already in use")
)
