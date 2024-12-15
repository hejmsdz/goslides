package main

import (
	"math"
	"strings"
)

type Measurer func(string) (float64, error)

func BreakLongLines(lines []string, measure Measurer, contentWidth float64) []string {
	result := make([]string, 0)

	for _, line := range lines {
		trimmedLine := strings.Trim(line, " ")
		escapedLine := preventAwkwardLineBreaks(trimmedLine)

		if !strings.Contains(escapedLine, " ") {
			escapedLine = trimmedLine
		}

		for _, wordSplit := range BreakOnSpaces(escapedLine, measure, contentWidth) {
			wordSplit = strings.ReplaceAll(wordSplit, "~", " ")
			result = append(result, wordSplit)
		}
	}

	return result
}

const LineEndMark = "\u200d"

var possiblePageBreakMarkers = map[string]int{
	"":  1,
	",": 1,
	":": 1,
	";": 1,
	".": 2,
	"?": 2,
	"!": 2,
}

func removeLineEndMarks(lines []string) []string {
	result := make([]string, len(lines))
	for i, line := range lines {
		if cutLine, didCut := strings.CutSuffix(line, LineEndMark); didCut {
			result[i] = cutLine
		} else {
			result[i] = line
		}
	}
	return result
}

func countFullLines(lines []string) int {
	numFullLines := 0
	for _, line := range lines {
		if strings.HasSuffix(line, LineEndMark) {
			numFullLines++
		}
	}
	return numFullLines
}

func SplitLongSlide(lines []string, maxLines int) [][]string {
	result := make([][]string, 0)
	numLines := len(lines)

	if numLines <= maxLines {
		result = append(result, removeLineEndMarks(lines))
		return result
	}

	for len(lines) > 0 {
		lineIndexToBreakOn := min(maxLines, len(lines)) - 1
		currentLinePriority := 0
		for i := lineIndexToBreakOn; i > 0; i-- {
			for marker, priority := range possiblePageBreakMarkers {
				if priority > currentLinePriority && strings.HasSuffix(lines[i], marker+LineEndMark) {
					// don't leave a single line (original, before line breaking) on the next slide,
					// unless a single line would remain on the slide before
					remainingLines := lines[i+1:]
					if len(remainingLines) == 0 || countFullLines(remainingLines) > 1 || countFullLines(lines[0:i]) == 1 {
						lineIndexToBreakOn = i
						currentLinePriority = priority
					}
				}
			}
		}
		endIdx := lineIndexToBreakOn + 1
		linesSlice := lines[0:endIdx]
		result = append(result, removeLineEndMarks(linesSlice))
		lines = lines[endIdx:]
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
			if len(word) <= 3 || (i == penultimateIndex && len(words[lastIndex]) < 10) {
				result += "~"
			} else {
				result += " "
			}
		}
	}

	return result
}

func BreakOnSpaces(line string, measure Measurer, contentWidth float64) []string {
	return breakLine(line+LineEndMark, measure, contentWidth, " ")
}

func breakLine(line string, measure Measurer, contentWidth float64, separator string) []string {
	result := make([]string, 0)

	lineWidth, _ := measure(line)
	if lineWidth <= contentWidth {
		result = append(result, line)
		return result
	}

	words := strings.SplitAfter(line, separator)
	start := 0
	end := len(words)

	for start < end {
		fragment := strings.Join(words[start:end], "")
		fragment = strings.Trim(fragment, " ")
		fragmentWidth, _ := measure(fragment)
		remainingFragmentWidth := lineWidth - fragmentWidth

		canBreakInTwoLines := start == 0 && fragmentWidth <= contentWidth && remainingFragmentWidth <= contentWidth

		if canBreakInTwoLines {
			return breakLineInTwo(words, measure, contentWidth, separator)
		}

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

func breakLineInTwo(words []string, measure Measurer, contentWidth float64, separator string) []string {
	firstLine := ""
	secondLine := strings.Join(words, "")
	minDiff := math.Inf(+1)
	i := 0

	for ; i < len(words); i++ {
		firstLine += separator + words[i]
		firstLine = strings.Trim(firstLine, " ")
		firstLineWidth, _ := measure(firstLine)

		secondLine = secondLine[len(words[i]):]
		secondLine = strings.Trim(secondLine, " ")
		secondLineWidth, _ := measure(secondLine)

		diff := math.Abs(firstLineWidth - secondLineWidth)
		if diff < minDiff {
			minDiff = diff
		} else {
			firstLine = firstLine[0 : len(firstLine)-len(words[i])]
			secondLine = words[i] + secondLine
			break
		}
	}

	return []string{firstLine, secondLine}
}
