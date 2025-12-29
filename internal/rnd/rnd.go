package rnd

import (
	"crypto/rand"
	unsafeRand "math/rand/v2"
)

//nolint:gochecknoglobals // нельзя сделать массив константой
var byteChoice = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// NotSafeGenRandomString НЕ ИСПОЛЬЗОВАТЬ ТАМ, ГДЕ НУЖЕН КРИПТОСТОЙКИЙ РАНДОМ (НАПРИМЕР, ДЛЯ STATE И PKCE).
func NotSafeGenRandomString(length uint) []byte {
	b := make([]byte, length)
	for i := range b {
		b[i] = byteChoice[unsafeRand.IntN(len(byteChoice))] //nolint:gosec // не тратимся на сисколы там, где не надо
	}
	return b
}

func GenRandomString(length uint) (string, error) {
	result := make([]byte, length)

	randomBytes := make([]byte, length*2)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	var i uint = 0
	for ; i < length; i++ {
		randomValue := int(randomBytes[i*2])<<8 + int(randomBytes[i*2+1])
		index := randomValue % len(byteChoice)
		result[i] = byteChoice[index]
	}

	return string(result), nil
}
