package db

import "errors"

var (
	ErrAccountAlreadyExists = errors.New("account already exists")
	ErrTokenNotFound        = errors.New("token not found")
)
