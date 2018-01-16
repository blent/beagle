package tracking

import "github.com/pkg/errors"

var (
	ErrStart = errors.New("tracker is already started")
	ErrStop  = errors.New("tracker is already stopped")
)
