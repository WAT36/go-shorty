package shortener

import (
	"crypto/rand"
	"errors"
	"math/big"
	"regexp"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const defaultLen = 6

var codeRe = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,32}$`)

func RandomCode() (string, error) {
	return RandomCodeN(defaultLen)
}

func RandomCodeN(n int) (string, error) {
	if n <= 0 {
		n = defaultLen
	}
	var out = make([]byte, n)
	for i := 0; i < n; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		out[i] = alphabet[idx.Int64()]
	}
	return string(out), nil
}

func ValidateCustom(code string) error {
	if !codeRe.MatchString(code) {
		return errors.New("invalid custom code (use 3-32 chars: a-zA-Z0-9_-) ")
	}
	return nil
}
