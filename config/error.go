package config

import "errors"

var (
	ErrorOutOfDated = errors.New("token is outdated")
	ErrorOutOfSpace = errors.New("space is not enough")
)
