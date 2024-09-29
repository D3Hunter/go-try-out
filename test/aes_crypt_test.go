package test

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

type Claims struct {
	Email      string
	ExpireTime int64
}

type AESCrypt struct {
	Block      cipher.Block
	CipherText []byte
}

func NewAESCrypt(cipherText []byte) (*AESCrypt, error) {
	block, err := aes.NewCipher(cipherText)
	if err != nil {
		return nil, err
	}

	return &AESCrypt{
		Block:      block,
		CipherText: cipherText,
	}, nil
}

func (s *AESCrypt) Encrypt(text []byte) []byte {
	return s.EncryptWithIv(text, s.CipherText[:s.Block.BlockSize()])
}

// https://gist.github.com/yingray/57fdc3264b1927ef0f984b533d63abab
func (s *AESCrypt) EncryptWithIv(text []byte, iv []byte) []byte {
	text = s.pkcs5Padding(text, s.Block.BlockSize())
	crypted := make([]byte, len(text))
	encryptor := cipher.NewCBCEncrypter(s.Block, iv)
	encryptor.CryptBlocks(crypted, text)
	return crypted
}

func (s *AESCrypt) Decrypt(crypted []byte) []byte {
	return s.DecryptWithIv(crypted, s.CipherText[:s.Block.BlockSize()])
}

func (s *AESCrypt) DecryptWithIv(crypted []byte, iv []byte) []byte {
	text := make([]byte, len(crypted))
	decryptor := cipher.NewCBCDecrypter(s.Block, iv)
	decryptor.CryptBlocks(text, crypted)
	return s.pkcs5UnPadding(text)
}

func (s *AESCrypt) pkcs5Padding(text []byte, blockSize int) []byte {
	padding := blockSize - len(text)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(text, padtext...)
}

func (s *AESCrypt) pkcs5UnPadding(cryptText []byte) []byte {
	length := len(cryptText)
	unpadding := int(cryptText[length-1])
	return cryptText[:(length - unpadding)]
}

const tokenPreFix = "CLOUD"

func TestEncrypt(t *testing.T) {
	aesCrypt, err := NewAESCrypt([]byte("1234567890123456"))
	expireTime := time.Now().Add(100 * 365 * 24 * time.Hour).Unix()
	claim := Claims{
		Email:      "test@pingcap.com",
		ExpireTime: expireTime,
	}
	data, err := json.Marshal(&claim)
	if err != nil {
		panic(err)
	}
	fmt.Println(tokenPreFix + base64.URLEncoding.EncodeToString(aesCrypt.Encrypt(data)))
}
