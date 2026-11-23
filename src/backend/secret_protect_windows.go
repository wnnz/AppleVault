//go:build windows

package backend

import (
	"encoding/base64"
	"unsafe"

	"golang.org/x/sys/windows"
)

var keychainProtectEntropy = []byte("AppleVault.keychain.v1")

func secretProtectionSupported() bool {
	return true
}

func protectSecret(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	raw := []byte(plain)
	in := windows.DataBlob{Size: uint32(len(raw)), Data: &raw[0]}
	entropy := windows.DataBlob{Size: uint32(len(keychainProtectEntropy)), Data: &keychainProtectEntropy[0]}
	var out windows.DataBlob
	if err := windows.CryptProtectData(&in, nil, &entropy, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out); err != nil {
		return "", err
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))
	encrypted := append([]byte(nil), unsafe.Slice(out.Data, out.Size)...)
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func unprotectSecret(encoded string) (string, error) {
	if encoded == "" {
		return "", nil
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	if len(raw) == 0 {
		return "", nil
	}
	in := windows.DataBlob{Size: uint32(len(raw)), Data: &raw[0]}
	entropy := windows.DataBlob{Size: uint32(len(keychainProtectEntropy)), Data: &keychainProtectEntropy[0]}
	var out windows.DataBlob
	if err := windows.CryptUnprotectData(&in, nil, &entropy, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out); err != nil {
		return "", err
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))
	decrypted := append([]byte(nil), unsafe.Slice(out.Data, out.Size)...)
	return string(decrypted), nil
}
