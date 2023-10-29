package random

import (
	"math/rand"
	"time"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func NewRandomString(length int) string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rnd.Intn(len(letterRunes))]
	}

	return string(b)
}
