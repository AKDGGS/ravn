package main

import (
	"regexp"
)

var YearAB_rx *regexp.Regexp = regexp.MustCompile(`(?:(\d{4})([a-z]{0,1})[;, ]*)+`)
var Nupper_rx *regexp.Regexp = regexp.MustCompile(`([A-Z]+)`)

func Levenshtein(str1, str2 string) int {
	// Convert string parameters to rune arrays to be compatible with non-ASCII
	runeStr1 := []rune(str1)
	runeStr2 := []rune(str2)

	// Get and store length of these strings
	runeStr1len := len(runeStr1)
	runeStr2len := len(runeStr2)
	if runeStr1len == 0 {
		return runeStr2len
	} else if runeStr2len == 0 {
		return runeStr1len
	} else if eq(runeStr1, runeStr2) {
		return 0
	}

	column := make([]int, runeStr1len+1)

	for y := 1; y <= runeStr1len; y++ {
		column[y] = y
	}
	for x := 1; x <= runeStr2len; x++ {
		column[0] = x
		lastkey := x - 1
		for y := 1; y <= runeStr1len; y++ {
			oldkey := column[y]
			var i int
			if runeStr1[y-1] != runeStr2[x-1] {
				i = 1
			}
			column[y] = min(
				min(column[y]+1, // insert
					column[y-1]+1), // delete
				lastkey+i) // substitution
			lastkey = oldkey
		}
	}

	return column[runeStr1len]
}

func min(a int, b int) int {
	if b < a {
		return b
	}
	return a
}

func eq(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func appendMap(m map[string]interface{}, k string, v interface{}) {
	switch t := v.(type) {
	case string:
		if a, ok := m[k].([]string); ok {
			m[k] = append(a, t)
		} else if s, ok := m[k].(string); ok {
			m[k] = []string{s, t}
		} else {
			m[k] = t
		}
	case int:
		if a, ok := m[k].([]int); ok {
			m[k] = append(a, t)
		} else if i, ok := m[k].(int); ok {
			m[k] = []int{i, t}
		} else {
			m[k] = t
		}
	}
}
