package app

import (
	"strings"
)

type ParseErrors struct {
	errors []error
}

func (e *ParseErrors) add(err error) {
	e.errors = append(e.errors, err)
}

func (e *ParseErrors) Error() string {
	var sb strings.Builder
	sb.WriteString("parser errors:\r\n")

	for _, err := range e.errors {
		sb.WriteString(err.Error())
		sb.WriteString("\r\n ")
	}
	return sb.String()
}

func (e *ParseErrors) get() error {
	if len(e.errors) == 0 {
		return nil
	}
	return e
}
