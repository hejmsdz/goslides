package core

import (
	"fmt"
	"strings"
)

const rows = 8
const cols = 20

func Tugalize(textDeck [][]string) string {
	slides := ""

	slides += pageHeader(0)
	slides += emptySlide()

	slideNo := 1
	for _, song := range textDeck {
		for _, verse := range song {
			verse = strings.ToUpper(verse)
			for _, slide := range textSlides(verse) {
				slides += pageHeader(slideNo)
				slides += slide
				slideNo++
			}
		}

		slides += pageHeader(slideNo)
		slides += emptySlide()
		slideNo++
	}

	return slides

	// encoder := charmap.Windows1250.NewEncoder()
	// encodedBytes, _ := encoder.Bytes([]byte(slides))

	// os.WriteFile("1208.txt", encodedBytes, 0644)
}

func pageHeader(pageNumber int) string {
	arrows := "<------------------>"
	return fmt.Sprintf("%s\n<- Strona nr: %03d ->\n%s\n", arrows, pageNumber, arrows)
}

func emptySlide() string {
	emptyLine := strings.Repeat(" ", cols) + "\n"
	asterisksLine := "*" + strings.Repeat(" ", cols-2) + "*" + "\n"
	slide := ""

	for i := 0; i < rows; i++ {
		if i == 0 || i == rows-1 {
			slide += asterisksLine
		} else {
			slide += emptyLine
		}

	}

	return slide
}

func textSlides(text string) []string {
	slides := make([]string, 0)

	measureText := func(s string) (float64, error) {
		return float64(len(s)), nil
	}

	lines := strings.Split(text, "\n")
	brokenLines := BreakLongLines(lines, measureText, cols)
	subPages := SplitLongSlide(brokenLines, rows)

	for _, subPage := range subPages {
		currentSlide := strings.Join(subPage, "\n") + "\n"
		currentSlide += strings.Repeat("\n", rows-len(subPage))

		slides = append(slides, currentSlide)
	}

	return slides
}
