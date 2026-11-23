//go:build !windows && !darwin

package backend

import "errors"

func secretProtectionSupported() bool {
	return false
}

func protectSecret(plain string) (string, error) {
	return "", errors.New("secret protection unsupported")
}

func unprotectSecret(encoded string) (string, error) {
	return "", errors.New("secret protection unsupported")
}
