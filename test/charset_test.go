package test

import (
	"fmt"
	"testing"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode/utf32"
)

func TestCharset(t *testing.T) {
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
