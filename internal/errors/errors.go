package myerrors

import (
	"fmt"
	"strings"
)

type InsertConflictError struct {
	OriginalURL []string
	Err         error
}

func (ce *InsertConflictError) Error() string {
	return fmt.Sprintf("duplicated record: %s; %v", strings.Join(ce.OriginalURL[:], ","), ce.Err)
}

func NewInsertConflictError(URL []string, err error) error {
	return &InsertConflictError{
		OriginalURL: URL,
		Err:         err,
	}
}

func (ce *InsertConflictError) Unwrap() error {
	return ce.Err
}
