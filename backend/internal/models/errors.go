package models

import (
	"errors"
)

var (
	ErrNoRecord           = errors.New("models: no matching record found")
	ErrInvalidCredentials = errors.New("models: invalid credentials")
	ErrDuplicateEmail     = errors.New("models: duplicate email")
	ErrDuplicateLogin     = errors.New("models: duplicate login")
	ErrInvalidSearchTerm  = errors.New("models: invalid search term")
)
