package xrand

import (
	"math/rand/v2"
)

const (
	codeLetters = "0123456789"
	charCode    = "0123456789abcdefghijklmnopqrstuvwxyz"
)

func RandCode() string {
	b := make([]byte, 6)
	for i := range b {
		b[i] = codeLetters[rand.IntN(len(codeLetters))]
	}
	return string(b)
}

func RandCharCode() string {
	b := make([]byte, 16)
	for i := range b {
		b[i] = charCode[rand.IntN(len(charCode))]
	}
	return string(b)
}
