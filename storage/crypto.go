package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

var commonIV = []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}

func Sha256Sum(text string) (sum string) {

	hash := sha256.New()
	hash.Write([]byte(text))
	return hex.EncodeToString(hash.Sum(nil))

}

func Encrypt(textString string, pass string) (result string, err error) {

	key := genKey(pass)

	c, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	text := []byte(textString)

	cfb := cipher.NewCFBEncrypter(c, commonIV)
	ciphertext := make([]byte, len(text))
	cfb.XORKeyStream(ciphertext, text)

	result = hex.EncodeToString(ciphertext)

	return

}

func genKey(pass string) (key string) {
	times := 24 / len(pass)

	key = strings.Repeat(pass, times+1)

	return key[0:24]
}

func Decrypt(textHex string, pass string) (result string, err error) {

	key := genKey(pass)
	c, err := aes.NewCipher([]byte(key))

	text, err := hex.DecodeString(textHex)
	if err != nil {
		return "", err
	}

	cfbdec := cipher.NewCFBDecrypter(c, commonIV)

	plaintext := make([]byte, len(text))
	cfbdec.XORKeyStream(plaintext, text)

	return string(plaintext), nil
}
