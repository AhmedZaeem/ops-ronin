package sanitize

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrEmptyInput      = errors.New("input is empty")
	ErrInvalidChars    = errors.New("input contains invalid characters")
	ErrPathTraversal   = errors.New("path traversal detected")
	ErrShellMetachars  = errors.New("input contains shell metacharacters")
	ErrControlChars    = errors.New("input contains control characters")
)

var shellMetachars = regexp.MustCompile(`[;&|<>\$\(\)\{\}\[\]\*\?\` + "`" + `\\!\n\r]`)
var controlChars = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)
var identifierRegex = regexp.MustCompile(`^[a-zA-Z0-9_.:-]+$`)
var argumentRegex = regexp.MustCompile(`^[a-zA-Z0-9_.:/=,@%~\-]+$`)

func Identifier(input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", ErrEmptyInput
	}

	if controlChars.MatchString(input) {
		return "", ErrControlChars
	}

	if strings.Contains(input, "..") {
		return "", ErrPathTraversal
	}

	if !identifierRegex.MatchString(input) {
		return "", ErrInvalidChars
	}

	if shellMetachars.MatchString(input) {
		return "", ErrShellMetachars
	}

	return input, nil
}

func Argument(input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", ErrEmptyInput
	}

	if controlChars.MatchString(input) {
		return "", ErrControlChars
	}

	if strings.Contains(input, "..") {
		return "", ErrPathTraversal
	}

	if !argumentRegex.MatchString(input) {
		return "", ErrInvalidChars
	}

	if shellMetachars.MatchString(input) {
		return "", ErrShellMetachars
	}

	return input, nil
}

func ContainerName(input string) (string, error) {
	clean, err := Identifier(input)
	if err != nil {
		return "", err
	}

	nameRegex := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]*$`)
	if !nameRegex.MatchString(clean) {
		return "", ErrInvalidChars
	}

	return clean, nil
}

func SocketPath(input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", ErrEmptyInput
	}

	if controlChars.MatchString(input) {
		return "", ErrControlChars
	}

	if strings.Contains(input, "..") {
		return "", ErrPathTraversal
	}

	if !strings.HasPrefix(input, "/") {
		return "", ErrInvalidChars
	}

	return input, nil
}
