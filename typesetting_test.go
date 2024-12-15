package main

import (
	"testing"
)

func TestBreakLongLines(t *testing.T) {
	longLine := "No way to cut it in two"
	measureText := func(s string) (float64, error) {
		return float64(len(s)), nil
	}
	contentWidth := 16.0
	result := BreakLongLines([]string{longLine}, measureText, contentWidth)

	if len(result) != 2 {
		t.Errorf("Expected the line to be broken into 2 lines")
	}
}

func TestBreakOnSpaces(t *testing.T) {
	longLine := "Chleb niebiański dał nam Pan"
	measureText := func(s string) (float64, error) {
		return float64(len(s)), nil
	}
	contentWidth := 25.0

	result := BreakOnSpaces(longLine, measureText, contentWidth)
	if len(result) != 2 {
		t.Errorf("Expected the line to be broken into 2 lines")
	}
	if result[0] != "Chleb niebiański" || result[1] != "dał nam Pan"+LineEndMark {
		t.Errorf("Unexpected line break")
	}
}

var lines = []string{
	"Lorem" + LineEndMark,
	"Ipsum" + LineEndMark,
	"Dolor." + LineEndMark,
	"Sit" + LineEndMark,
	"Amet," + LineEndMark,
	"Consectetur?" + LineEndMark,
	"Adipiscing" + LineEndMark,
	"Elit!" + LineEndMark,
}

func TestSplitLongSlide5(t *testing.T) {
	result := SplitLongSlide(lines, 5)

	if len(result) != 2 {
		t.Errorf("Expected the slide to be split into 2 pages")
	}

	if result[1][0] != "Sit" {
		t.Errorf("Expected the break to happen after the period (.)")
	}
}

func TestSplitLongSlide4(t *testing.T) {
	result := SplitLongSlide(lines, 4)

	if len(result) != 3 {
		t.Errorf("Expected the slide to be split into 3 pages")
	}

	if result[1][0] != "Sit" {
		t.Errorf("Expected the first break to happen after the period (.)")
	}

	if result[2][0] != "Adipiscing" {
		t.Errorf("Expected the break to happen after the question mark (?)")
	}
}

func TestSplitLongSlide3(t *testing.T) {
	result := SplitLongSlide(lines, 3)

	if len(result) != 3 {
		t.Errorf("Expected the slide to be split into 3 pages")
	}

	if result[1][0] != "Sit" {
		t.Errorf("Expected the first break to happen after the period (.)")
	}

	if result[2][0] != "Adipiscing" {
		t.Errorf("Expected the break to happen after the question mark (?)")
	}
}
