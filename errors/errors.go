package errors

import (
	"errors"
)

var (
	// ErrNotSupport not support this function
	ErrNotSupport = errors.New("not support this function yet")
	// ErrCanceled canceled
	ErrCanceled = errors.New("the context had been canceled")
	// ErrFailedToListen cannot listen on an address, or bad port
	ErrFailedToListen = errors.New("failed to listen, bad port")
)
