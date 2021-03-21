package main

import (
	"strings"
)

type Measurer func(string) (float64, error)

func BreakLongLines(lines []string, measure Measurer, contentWidth float64) []string {
	result := make([]string, 0)

	for _, line := range lines {
		line = strings.Trim(line, " ")
		line = preventAwkwardLineBreaks(line)

		for _, wordSplit := range breakOnSpaces(line, measure, contentWidth) {
			wordSplit = strings.ReplaceAll(wordSplit, "~", " ")
			result = append(result, wordSplit)
		}
	}

	return result
}

func preventAwkwardLineBreaks(text string) string {
	result := ""
	words := strings.Split(text, " ")
	lastIndex := len(words) - 1
	penultimateIndex := lastIndex - 1

	for i, word := range words {
		result += word
		if i < lastIndex {
			if len(word) <= 3 || i == penultimateIndex {
				result += "~"
			} else {
				result += " "
			}
		}
	}

	return result
}

func breakOnSpaces(line string, measure Measurer, contentWidth float64) []string {
	return breakLine(line, measure, contentWidth, " ")
}

func breakLine(line string, measure Measurer, contentWidth float64, separator string) []string {
	result := make([]string, 0)

	if lineWidth, _ := measure(line); lineWidth <= contentWidth {
		result = append(result, line)
		return result
	}

	var words []string
	words = strings.SplitAfter(line, separator)

	start := 0
	end := len(words)

	for start < end {
		fragment := strings.Join(words[start:end], "")
		fragment = strings.Trim(fragment, " ")
		fragmentWidth, _ := measure(fragment)
		if fragmentWidth <= contentWidth {
			result = append(result, fragment)
			start = end
			end = len(words)
		} else {
			end--
		}
	}
	if end != len(words) {
		for ; end < len(words); end++ {
			result = append(result, words[end])
		}
	}

	return result
}
