package utils

import (
	"crypto/rand"
	"math/big"
)

var (
	letters    = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890*-.")
	lettersLen = len(letters)
)

func RandString(length int) string {
	r := make([]rune, length)
	for i := range r {
		pos, _ := rand.Int(rand.Reader, big.NewInt(int64(lettersLen)))
		// Some ENVs as passwords cannot start with a non-letter
		if i == 0 {
			pos, _ = rand.Int(rand.Reader, big.NewInt(int64(26*2)))
		}
		r[i] = letters[pos.Int64()]
	}

	return string(r)
}
