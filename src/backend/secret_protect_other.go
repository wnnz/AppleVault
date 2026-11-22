//go:build !windows

package backend

import "errors"

func protectSecret(plain string) (string, error) {
	return "", errors.New("secret protection unsupported")
}

func unprotectSecret(encoded string) (string, error) {
	return "", errors.New("secret protection unsupported")
}
