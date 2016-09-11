package main

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

func transformInt(line string) string {
	tokens := strings.Split(line, " ")
	var transformedTokens []string
	for _, token := range tokens {
		num, err := strconv.Atoi(token)
		if err != nil {
			transformedTokens = append(transformedTokens, token)
		} else {
			transformedTokens = append(transformedTokens, strconv.Itoa(num+123))
		}
	}

	return strings.Join(transformedTokens, " ")
}

func transformString(line string) string {
	tokens := strings.Split(line, " ")
	var transformedTokens []string
	for _, token := range tokens {
		size := utf8.RuneCountInString(token)
		reversed := make([]rune, size)
		chCount := 1
		for _, ch := range token {
			if unicode.IsLower(ch) {
				reversed[size-chCount] = unicode.ToUpper(ch)
			} else if unicode.IsUpper(ch) {
				reversed[size-chCount] = unicode.ToLower(ch)
			} else {
				reversed[size-chCount] = ch
			}
			chCount++
		}
		transformedTokens = append(transformedTokens, string(reversed))
	}
	return strings.Join(transformedTokens, " ")
}

func noTransform(line string) string {
	return line
}
