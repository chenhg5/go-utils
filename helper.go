package utils

import (
	"time"
	"math/rand"
	"github.com/NebulousLabs/fastrand"
)

func Uuid(length int64) string {
	ele := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0", "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "v", "k",
		"l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "A", "B", "C", "D", "E", "F", "G",
		"H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
	ele, _ = Random(ele)
	uuid := ""
	var i int64
	for i = 0; i < length; i++ {
		rand.Seed(time.Now().Unix() + i)
		uuid += ele[fastrand.Intn(59)]
	}
	return uuid
}

func Random(strings []string) ([]string, error) {
	for i := len(strings) - 1; i > 0; i-- {
		num := fastrand.Intn(i + 1)
		strings[i], strings[num] = strings[num], strings[i]
	}

	str := make([]string, 0)
	for i := 0; i < len(strings); i++ {
		str = append(str, strings[i])
	}
	return str, nil
}