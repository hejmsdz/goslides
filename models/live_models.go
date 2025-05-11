package models

import "time"

type LiveSession struct {
	URL         string
	CurrentPage int
	Token       string
	FileName    string
	UpdatedAt   time.Time
}
