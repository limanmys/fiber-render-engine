package helpers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
)

type laravelAesKey struct {
	Iv    []byte `json:"iv"`
	Value []byte `json:"value"`
	Mac   string `json:"mac"`
	Tag   string `json:"tag"`
}

func LaravelAesDecrypt(name, payload string) (string, error) {
	cipherText, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return "", err
	}

	aesKey := laravelAesKey{}
	err = json.Unmarshal(cipherText, &aesKey)
	if err != nil {
		return "", err
	}

	dummykey := Env("APP_KEY", "")
	dummykey = dummykey[7:]

	key, err := base64.StdEncoding.DecodeString(dummykey)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(aesKey.Value)%aes.BlockSize != 0 {
		return "", errors.New("block size cant be zero")
	}

	mode := cipher.NewCBCDecrypter(block, aesKey.Iv)
	mode.CryptBlocks(aesKey.Value, aesKey.Value)
	plainText := string(unpadding(aesKey.Value))

	validator := createValidator(name, key)
	if strings.HasPrefix(plainText, validator) {
		plainText = strings.TrimPrefix(plainText, validator)
	} else {
		return "", errors.New("token cannot be validated")
	}

	return plainText, nil
}

func unpadding(src []byte) []byte {
	length := len(src)
	unpadding := int(src[length-1])

	return src[:(length - unpadding)]
}

func createValidator(cookieName string, key []byte) string {
	h := hmac.New(sha1.New, key)
	h.Write([]byte(cookieName + "v2"))
	return hex.EncodeToString(h.Sum(nil)) + "|"
}
