package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image/png"
	"io"

	"github.com/gen2brain/go-fitz"
)

func BuildImages(textDeck [][]string, pageConfig PageConfig) (io.Reader, []ContentSlide, error) {
	pdf, contents, err := BuildPDF(textDeck, pageConfig)
	if err != nil {
		return nil, nil, err
	}
	defer pdf.Close()

	doc, err := fitz.NewFromMemory(pdf.GetBytesPdf())
	if err != nil {
		return nil, nil, err
	}
	defer doc.Close()

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	defer zipWriter.Close()

	for n := 0; n < doc.NumPage(); n++ {
		img, err := doc.Image(n)
		if err != nil {
			return nil, nil, err
		}

		name := fmt.Sprintf("%03d.png", n)
		f, err := zipWriter.Create(name)
		if err != nil {
			return nil, nil, err
		}

		err = png.Encode(f, img)
		if err != nil {
			return nil, nil, err
		}
	}

	return buf, contents, nil
}
