package main

import (
	"testing"
)

var lines = []string{
	"Lorem",
	"Ipsum",
	"Dolor.",
	"Sit",
	"Amet,",
	"Consectetur?",
	"Adipiscing",
	"Elit!",
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
