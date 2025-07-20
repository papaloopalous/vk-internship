package encryption

import (
	"api/internal/logger"
	"api/internal/messages"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strconv"
)

// deriveKeyAndIV генерирует ключ и вектор инициализации из основного ключа и соли
func deriveKeyAndIV(key, salt []byte) (k32, iv16 []byte) {
	var buf []byte
	prev := key
	for len(buf) < 48 {
		d := md5.Sum(append(prev, salt...))
		buf = append(buf, d[:]...)
		prev = d[:]
	}
	return buf[:32], buf[32:48]
}

// pkcs7Unpad удаляет PKCS7 паддинг
func pkcs7Unpad(b []byte) ([]byte, error) {
	if len(b) == 0 {
		return nil, errors.New(messages.LogErrEmptyData)
	}

	pad := int(b[len(b)-1])
	if pad == 0 || pad > aes.BlockSize {
		return nil, errors.New(messages.LogErrPadding)
	}

	for i := len(b) - pad; i < len(b); i++ {
		if b[i] != byte(pad) {
			return nil, errors.New(messages.LogErrPadding)
		}
	}
	return b[:len(b)-pad], nil
}

// DecryptData расшифровывает данные с использованием AES-CBC
func DecryptData(cipherB64, sharedKeyHex string) (string, error) {
	keyBytes, err := hex.DecodeString(sharedKeyHex)
	if err != nil {
		logger.Error(messages.ServiceEncryption, messages.LogErrHexDecode, map[string]string{
			messages.LogDetails: err.Error(),
		})
		return "", errors.New(messages.ClientErrDecryption)
	}

	if len(keyBytes) != messages.CryptoKeyLength {
		logger.Error(messages.ServiceEncryption, messages.LogErrKeyLength, map[string]string{
			messages.LogExpected: strconv.Itoa(messages.CryptoKeyLength),
			messages.LogGot:      strconv.Itoa(len(keyBytes)),
		})
		return "", errors.New(messages.ClientErrDecryption)
	}

	raw, err := base64.StdEncoding.DecodeString(cipherB64)
	if err != nil {
		logger.Error(messages.ServiceEncryption, messages.LogErrBase64Decode, map[string]string{
			messages.LogDetails: err.Error(),
		})
		return "", errors.New(messages.ClientErrDecryption)
	}

	if len(raw) < 16 || string(raw[:8]) != messages.CryptoSaltedPrefix {
		logger.Error(messages.ServiceEncryption, messages.LogErrMissingSalt, nil)
		return "", errors.New(messages.ClientErrDecryption)
	}

	salt, ciphertext := raw[8:16], raw[16:]
	key, iv := deriveKeyAndIV(keyBytes, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		logger.Error(messages.ServiceEncryption, messages.LogErrCipherInit, map[string]string{
			messages.LogDetails: err.Error(),
		})
		return "", errors.New(messages.ClientErrDecryption)
	}

	if len(ciphertext)%aes.BlockSize != 0 {
		logger.Error(messages.ServiceEncryption, messages.LogErrBlockSize, map[string]string{
			messages.LogBlockSize: strconv.Itoa(aes.BlockSize),
		})
		return "", errors.New(messages.ClientErrDecryption)
	}

	plain := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plain, ciphertext)

	plain, err = pkcs7Unpad(plain)
	if err != nil {
		logger.Error(messages.ServiceEncryption, messages.LogErrPadding, map[string]string{
			messages.LogDetails: err.Error(),
		})
		return "", errors.New(messages.ClientErrDecryption)
	}

	logger.Info(messages.ServiceEncryption, messages.LogStatusDecryption, map[string]string{
		messages.LogLength: strconv.Itoa(len(plain)),
	})
	return string(plain), nil
}

// pkcs7Pad добавляет PKCS7 паддинг
func pkcs7Pad(b []byte) []byte {
	p := aes.BlockSize - len(b)%aes.BlockSize
	for i := 0; i < p; i++ {
		b = append(b, byte(p))
	}
	return b
}

// EncryptData шифрует данные с использованием AES-CBC
func EncryptData(plaintext, sharedKeyHex string) (string, error) {
	keyBytes, err := hex.DecodeString(sharedKeyHex)
	if err != nil {
		logger.Error(messages.ServiceEncryption, messages.LogErrHexDecode, map[string]string{
			messages.LogDetails: err.Error(),
		})
		return "", errors.New(messages.ClientErrEncryption)
	}

	if len(keyBytes) != messages.CryptoKeyLength {
		logger.Error(messages.ServiceEncryption, messages.LogErrKeyLength, map[string]string{
			messages.LogExpected: strconv.Itoa(messages.CryptoKeyLength),
			messages.LogGot:      strconv.Itoa(len(keyBytes)),
		})
		return "", errors.New(messages.ClientErrEncryption)
	}

	sum := sha256.Sum256(append(keyBytes, []byte(plaintext)...))
	salt := sum[:messages.CryptoSaltLength]

	k, iv := deriveKeyAndIV(keyBytes, salt)

	block, err := aes.NewCipher(k)
	if err != nil {
		logger.Error(messages.ServiceEncryption, messages.LogErrCipherInit, map[string]string{
			messages.LogDetails: err.Error(),
		})
		return "", errors.New(messages.ClientErrEncryption)
	}

	plain := pkcs7Pad([]byte(plaintext))
	ciphertext := make([]byte, len(plain))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext, plain)

	out := append([]byte(messages.CryptoSaltedPrefix), salt...)
	out = append(out, ciphertext...)

	logger.Info(messages.ServiceEncryption, messages.LogStatusEncryption, map[string]string{
		messages.LogLength: strconv.Itoa(len(out)),
	})
	return base64.StdEncoding.EncodeToString(out), nil
}
