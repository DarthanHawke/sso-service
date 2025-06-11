package ssoerrors

import "errors"

var (
	ErrNotFound = errors.New("not found")
	ErrInternal = errors.New("internal error")
)

var (
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrPasswordTooWeak = errors.New("password must be at least 8 characters long, contain uppercase, lowercase, digit and special character")
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrInvalidToken    = errors.New("accsess data are outdated")
)

var (
	ErrRoleExists      = errors.New("role already exists")
	ErrRoleNotFound    = errors.New("role not found")
	ErrRoleNotAssigned = errors.New("role is not assigned to the user")
)
