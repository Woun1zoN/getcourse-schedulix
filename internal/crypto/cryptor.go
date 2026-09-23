package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
)

type Cryptor struct {
    aead cipher.AEAD
}

func NewCryptor(key []byte) (*Cryptor, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    return &Cryptor{aead: gcm}, nil
}

func (c *Cryptor) Encrypt(plaintext string) (string, error) {
    nonce := make([]byte, c.aead.NonceSize())

    if _, err := rand.Read(nonce); err != nil {
        return "", err
    }

    ciphertext := c.aead.Seal(nonce, nonce, []byte(plaintext), nil)

    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (c *Cryptor) Decrypt(encoded string) (string, error) {
    data, err := base64.StdEncoding.DecodeString(encoded)
    if err != nil {
        return "", err
    }

    nonceSize := c.aead.NonceSize()
    nonce, ciphertext := data[:nonceSize], data[nonceSize:]

    plaintext, err := c.aead.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return "", err
    }

    return string(plaintext), nil
}