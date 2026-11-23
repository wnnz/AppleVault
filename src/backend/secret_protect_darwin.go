//go:build darwin

package backend

import (
	"crypto/rand"
	"errors"
	"io"
	"sync"

	"github.com/byteness/keyring"
)

const (
	darwinSettingsKeyService = "com.wnnz.applevault.settings-key"
	darwinSettingsKeyName    = "master-v1"
)

var (
	darwinMasterKeyMu sync.Mutex
	darwinMasterKey   []byte
)

func secretProtectionSupported() bool {
	return true
}

func openDarwinSettingsKeyring() (keyring.Keyring, error) {
	return keyring.Open(keyring.Config{
		AllowedBackends:                []keyring.BackendType{keyring.KeychainBackend},
		ServiceName:                    darwinSettingsKeyService,
		KeychainTrustApplication:       true,
		KeychainAccessibleWhenUnlocked: true,
	})
}

func loadOrCreateDarwinMasterKey(createIfMissing bool) ([]byte, error) {
	darwinMasterKeyMu.Lock()
	defer darwinMasterKeyMu.Unlock()

	if len(darwinMasterKey) == 32 {
		return append([]byte(nil), darwinMasterKey...), nil
	}

	ring, err := openDarwinSettingsKeyring()
	if err != nil {
		return nil, err
	}
	if item, getErr := ring.Get(darwinSettingsKeyName); getErr == nil && len(item.Data) == 32 {
		darwinMasterKey = append([]byte(nil), item.Data...)
		return append([]byte(nil), darwinMasterKey...), nil
	} else if getErr != nil && !errors.Is(getErr, keyring.ErrKeyNotFound) {
		return nil, getErr
	} else if !createIfMissing {
		return nil, errors.New("darwin keychain master key not found")
	}

	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	if err := ring.Set(keyring.Item{
		Key:                       darwinSettingsKeyName,
		Data:                      key,
		Label:                     "AppleVault Settings Protection Key",
		KeychainNotSynchronizable: true,
	}); err != nil {
		return nil, err
	}
	darwinMasterKey = append([]byte(nil), key...)
	return key, nil
}

func protectSecret(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	key, err := loadOrCreateDarwinMasterKey(true)
	if err != nil {
		return "", err
	}
	return sealSecretAESGCM(key, plain)
}

func unprotectSecret(encoded string) (string, error) {
	if encoded == "" {
		return "", nil
	}
	key, err := loadOrCreateDarwinMasterKey(false)
	if err != nil {
		return "", err
	}
	return openSecretAESGCM(key, encoded)
}
