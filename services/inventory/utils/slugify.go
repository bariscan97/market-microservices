package utils

import (
	"strings"
	"unicode"
)

func Slugify(s string) string {

	var (
		result string
		flag   bool
	)
	
	for _, i := range s {

		if unicode.IsSpace(i) {
			flag = true
			
		}else {
			if len(result) > 0 && flag {
				result += "-"
			}
			result += strings.ToLower(string(i))
			flag = false
		}

		
	}

	return result
}
