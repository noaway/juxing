package utils

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"math/rand"
	"net/url"
	"runtime"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// GetMD5HashString func
func GetMD5HashString(str string) string {
	return GetMD5HashBytes([]byte(str))
}

// GetMD5HashBytes func
func GetMD5HashBytes(data []byte) string {
	hasher := md5.New()
	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil))
}

// GetRandomString func
func GetRandomString(n int, alphabets ...byte) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	_, err := rand.Read(bytes)
	if err != nil {
		return ""
	}
	for i, b := range bytes {
		if len(alphabets) == 0 {
			bytes[i] = alphanum[b%byte(len(alphanum))]
		} else {
			bytes[i] = alphabets[b%byte(len(alphabets))]
		}
	}
	return string(bytes)
}

// Errors func
func Errors(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

// WaitTimeout fn
func WaitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false
	case <-time.After(timeout):
		return true
	}
}

func DeferError(errorfn func(string, interface{}), dones ...func()) {
	if err := recover(); err != nil {
		var buf [2 << 10]byte
		errorfn(string(buf[:runtime.Stack(buf[:], false)]), err)
	}
	for _, done := range dones {
		done()
	}
}

// UTF8Reader struct
type UTF8Reader struct {
	buffer *bufio.Reader
}

func (rd *UTF8Reader) Read(b []byte) (n int, err error) {
	for {
		var r rune
		var size int
		r, size, err = rd.buffer.ReadRune()
		if err != nil {
			return
		}
		if r == unicode.ReplacementChar && size == 1 {
			continue
		} else if n+size < len(b) {
			utf8.EncodeRune(b[n:], r)
			n += size
		} else {
			rd.buffer.UnreadRune()
			break
		}
	}
	return
}

// NewUTF8Reader constructs a new ValidUTF8Reader that wraps an existing io.Reader
func NewUTF8Reader(rd io.Reader) *UTF8Reader {
	return &UTF8Reader{bufio.NewReader(rd)}
}

// GbkToUtf8 func
func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

// MustGbkToUtf8 fn
func MustGbkToUtf8(s []byte) []byte {
	d, _ := GbkToUtf8(s)
	return d
}

// Unescape fn
func Unescape(s string) (n string) {
	n, _ = url.QueryUnescape(s)
	return
}
