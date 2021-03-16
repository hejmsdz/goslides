package main

import (
	"fmt"
	"os"
	"time"

	"github.com/signintech/gopdf"
)

func scheduleRemove(path string) {
	time.Sleep(60 * time.Second)
	os.Remove(path)
}

func SaveTemporaryPDF(pdf *gopdf.GoPdf, name string) {
	path := fmt.Sprintf("public/%s", name)
	pdf.WritePdf(path)
	go scheduleRemove(path)
}
