package gorep

import (
	"unicode"
)

type StringCaseConverter struct {
}

func (StringCaseConverter) SnakeCaseToCamelCase(input string) (result string) {
	isToUpper := false

	for i, letter := range input {
		if i == 0 {
			result = string(unicode.ToUpper(letter))
			continue
		}

		if isToUpper {
			result += string(unicode.ToUpper(letter))
			isToUpper = false
		} else {
			if letter == '_' {
				isToUpper = true
			} else {
				result += string(letter)
			}
		}
	}

	return
}

func (StringCaseConverter) Lowercase(input string) (result string) {
	letters := []rune(input)
	letters[0] = unicode.ToLower(letters[0])

	return string(letters)
}
