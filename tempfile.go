package main

import (
	"fmt"
	"io"
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

func SaveTemporaryFile(src io.Reader, name string) {
	path := fmt.Sprintf("public/%s", name)
	dest, err := os.Create(path)
	if err != nil {
		return
	}

	io.Copy(dest, src)
	go scheduleRemove(path)
}
