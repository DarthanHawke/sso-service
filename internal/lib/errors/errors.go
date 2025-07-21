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
	ErrSessionOld      = errors.New("session is old")
)

var (
	ErrRoleExists      = errors.New("role already exists")
	ErrRoleNotFound    = errors.New("role not found")
	ErrRoleName        = errors.New("role name cannot be empty")
	ErrRoleNotAssigned = errors.New("role is not assigned to the user")
	ErrRoleHasUsers    = errors.New("role has user")
)

var (
	ErrPermissionExists   = errors.New("permission already exists")
	ErrPermissionNotFound = errors.New("permission not found")
	ErrPermissionName     = errors.New("permission name cannot be empty")
	ErrPermissionHasRoles = errors.New("permission has user")
)
