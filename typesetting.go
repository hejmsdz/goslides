package main

import (
	"regexp"
	"strings"
)

type Measurer func(string) (float64, error)

var shortWordsRegexp = regexp.MustCompile(" ([\\p{L}~]{1,3}) ")
var finalWordRegexp = regexp.MustCompile(" (\\p{L}+[[:punct:]]?)$")

func BreakLongLines(lines []string, measure Measurer, contentWidth float64) []string {
	result := make([]string, 0)

	for _, line := range lines {
		line = strings.Trim(line, " ")
		line = shortWordsRegexp.ReplaceAllString(line, " $1~")
		line = finalWordRegexp.ReplaceAllString(line, "~$1")

		for _, wordSplit := range breakOnSpaces(line, measure, contentWidth) {
			wordSplit = strings.ReplaceAll(wordSplit, "~", " ")
			result = append(result, wordSplit)
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
