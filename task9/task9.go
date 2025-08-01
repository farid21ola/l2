package main

import (
	"errors"
	"fmt"
	"unicode"
)

func main() {
	a := "a4bc2d5e"
	b, err := Unpacking(a)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(b) // aaaabccddddde
}

func Unpacking(str string) (string, error) {
	res := make([]rune, 0, len(str))
	var last rune
	count := 0
	for i, v := range str {
		if unicode.IsDigit(v) {
			if i == 0 {
				return "", errors.New("invalid string")
			} else if last == '\\' {
				last = v
			} else {
				count = count*10 + int(v-'0')
			}
		} else if v == '\\' {
			if last != 0 {
				if count == 0 {
					res = append(res, last)
				} else {
					for j := 0; j < count; j++ {
						res = append(res, last)
					}
					count = 0
				}
			}
			last = '\\'
		} else {
			if last != 0 {
				if count == 0 {
					res = append(res, last)
				} else {
					for j := 0; j < count; j++ {
						res = append(res, last)
					}
					count = 0
				}
			}
			last = v
		}
	}
	if last != 0 {
		if count != 0 {
			for j := 0; j < count; j++ {
				res = append(res, last)
			}
		} else {
			res = append(res, last)
		}
	}
	return string(res), nil
}
