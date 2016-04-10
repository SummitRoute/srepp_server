////////////////////////////////////////////////////////////////////////////
//
// Summit Route End Point Protection
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.
//
/////////////////////////////////////////////////////////////////////////////


package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// KeyForObfuscation is a crypto key that should only be used for obfuscating data
var KeyForObfuscation = []byte("1sf2324knkfnsfff")

// SymmetricEncrypt to be used for small symmetric crypto operations
// Taken from: http://stackoverflow.com/questions/18817336/golang-encrypting-a-string-with-aes-and-base64
func SymmetricEncrypt(key []byte, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(plaintext))
	return ciphertext, nil
}

// SymmetricDecrypt to be used for small symmetric crypto operations
func SymmetricDecrypt(key []byte, input []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(input) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}

	iv := input[:aes.BlockSize]
	text := input[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)

	return text, nil
}
