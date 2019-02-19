package utils

import (
	"math"
	"math/rand"
	"time"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func RandomInt(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}

func RandomFloat64(min, max float64) float64 {
	f := min + rand.Float64()*(max-min)
	precision := 1
	return float64(int(f*math.Pow(10, float64(precision)))) / math.Pow(10, float64(precision))
}
