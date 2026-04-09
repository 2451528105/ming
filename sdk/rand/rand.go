package rand

import "math/rand"

func RandLittleLetter(x int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyz")
	b := make([]rune, x)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
