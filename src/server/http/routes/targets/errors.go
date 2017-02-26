package routes

import "github.com/pkg/errors"

var (
	ErrMissedId      = errors.New("missed target id")
	ErrInvalidId     = errors.New("invalid target id")
	ErrInvalidTarget = errors.New("invalid target")
)
