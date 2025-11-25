package service

import (
	"errors"
	"fmt"
)

var (
	ErrValidation = errors.New("validation error")
	ErrNotFound   = errors.New("resource not found")
)

func ValidationError(message string) error {
	return fmt.Errorf("%w: %s", ErrValidation, message)
}


