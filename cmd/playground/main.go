package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/http/pprof"
	"sort"
	"strconv"
	"time"

	"github.com/pingcap/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode/utf32"
)

type test struct {
	val int
}

var globalTest test

func (t test) checkCopyOnStruct() {
	fmt.Printf("val in global test: %d\n", globalTest.val)
	fmt.Printf("val in test: %d\n", t.val)
	t.val = 100
	fmt.Printf("val in global test: %d\n", globalTest.val)
	fmt.Printf("val in test: %d\n", t.val)
}
func (t *test) checkCopyOnPointer() {
	fmt.Printf("val in global test: %d\n", globalTest.val)
	fmt.Printf("val in test: %d\n", t.val)
	t.val = 100
	fmt.Printf("val in global test: %d\n", globalTest.val)
	fmt.Printf("val in test: %d\n", t.val)
}
func (t *test) checkNilThis() {
	if t == nil {
		fmt.Printf("t is nil\n")
	} else {
		fmt.Printf("t is not nil\n")
	}
}
func (t test) xx(i int) {
	t.val = i
}

func testCharset() {
	fmt.Println("\xbd\xb2\x3d\xbc\x20\xe2\x8c\x98")
	fmt.Println("\xd9\xf1")
	fmt.Println("\u00e0\u0300\u0061")
	fmt.Printf("%d\n", '\u2318')
	const r = rune(8984)
	fmt.Printf("%x %x %s\n", r, string(r), string(r))
	s := "\xFF"
	runes := []rune(s)
	for _, r := range runes {
		fmt.Printf("%x ", r)
	}
	fmt.Println()
	fmt.Println("\x00\x00\x23\x18")
	decoder := utf32.UTF32(utf32.BigEndian, utf32.IgnoreBOM).NewDecoder()
	s2, err := decoder.String("\x00\x00\x23\x18")
	fmt.Println(s2, err)
	newDecoder := simplifiedchinese.GB18030.NewDecoder()
	s3, err := newDecoder.String("\xd9\xf1")
	fmt.Println(s3, err)
}

type item struct {
	val int
}

func (i *item) String() string {
	return strconv.Itoa(i.val)
}

func returnResult() sql.Result {
	//v := driver.RowsAffected(0)
	//return v
	return nil
}

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

func encrypt() {
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

var testGauge = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "playground",
		Subsystem: "test",
		Name:      "gauge",
	},
)

// InitStatus initializes the HTTP status server.
func InitStatus(lis net.Listener) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	httpS := &http.Server{
		Handler: mux,
	}
	err := httpS.Serve(lis)
	if err != nil && err != http.ErrServerClosed {
		log.L().Error("status server returned", zap.Error(err))
	}
}

func testMetricmain() {
	prometheus.DefaultRegisterer.MustRegister(testGauge)
	go func() {
		for i := 0; i < 10000; i++ {
			time.Sleep(1 * time.Second)
			testGauge.Set(100 * rand.Float64())
		}
	}()
	rootLis, err := net.Listen("tcp", net.JoinHostPort("0.0.0.0", "8361"))
	if err != nil {
		panic(err)
	}
	InitStatus(rootLis)
}

func main() {
	var slice []int
	sort.Slice(slice, func(i, j int) bool {
		return slice[i] < slice[j]
	})
}
