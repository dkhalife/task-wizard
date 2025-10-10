package encryption

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt(t *testing.T) {
	encryptor := NewEncryptor("test-secret-key")

	tests := []struct {
		name      string
		plaintext string
	}{
		{
			name:      "simple text",
			plaintext: "hello world",
		},
		{
			name:      "token",
			plaintext: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ",
		},
		{
			name:      "base64 certificate data",
			plaintext: "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURJekNDQWd1Z0F3SUJBZ0lVVmVmTzU4VnZ=",
		},
		{
			name:      "empty string",
			plaintext: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			encrypted, err := encryptor.Encrypt(tt.plaintext)
			require.NoError(t, err)

			if tt.plaintext == "" {
				assert.Equal(t, "", encrypted)
				return
			}

			// Should not be the same as plaintext
			assert.NotEqual(t, tt.plaintext, encrypted)

			// Decrypt
			decrypted, err := encryptor.Decrypt(encrypted)
			require.NoError(t, err)
			assert.Equal(t, tt.plaintext, decrypted)
		})
	}
}

func TestEncryptDeterministic(t *testing.T) {
	encryptor := NewEncryptor("test-secret-key")
	plaintext := "test data"

	// Encrypt the same data twice
	encrypted1, err := encryptor.Encrypt(plaintext)
	require.NoError(t, err)

	encrypted2, err := encryptor.Encrypt(plaintext)
	require.NoError(t, err)

	// Should produce different ciphertexts (due to random nonce)
	assert.NotEqual(t, encrypted1, encrypted2)

	// But both should decrypt to the same plaintext
	decrypted1, err := encryptor.Decrypt(encrypted1)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted1)

	decrypted2, err := encryptor.Decrypt(encrypted2)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted2)
}

func TestDecryptInvalidData(t *testing.T) {
	encryptor := NewEncryptor("test-secret-key")

	tests := []struct {
		name       string
		ciphertext string
		wantErr    bool
	}{
		{
			name:       "invalid base64",
			ciphertext: "not-valid-base64!!!",
			wantErr:    true,
		},
		{
			name:       "too short ciphertext",
			ciphertext: "YWJj", // "abc" in base64 - too short
			wantErr:    true,
		},
		{
			name:       "empty string",
			ciphertext: "",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := encryptor.Decrypt(tt.ciphertext)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDifferentKeys(t *testing.T) {
	encryptor1 := NewEncryptor("key1")
	encryptor2 := NewEncryptor("key2")

	plaintext := "sensitive data"

	// Encrypt with key1
	encrypted, err := encryptor1.Encrypt(plaintext)
	require.NoError(t, err)

	// Try to decrypt with key2 (should fail)
	_, err = encryptor2.Decrypt(encrypted)
	assert.Error(t, err)
}
