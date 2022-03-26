package main

import (
	"strings"

	"github.com/signintech/gopdf"
	"github.com/skip2/go-qrcode"
)

const pageWidth float64 = 768
const pageHeight float64 = 576
const margin float64 = 50
const fontSize int = 36
const hintFontSize int = 24
const lineSpacing float64 = 1.3
const font string = "./fonts/source-sans-pro.ttf"

const contentWidth = pageWidth - 2*margin

func createNewPDF() (*gopdf.GoPdf, error) {
	pageSize := gopdf.Rect{W: pageWidth, H: pageHeight}

	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: pageSize})

	err := pdf.AddTTFFont("default", font)
	if err != nil {
		return nil, err
	}

	addPage(&pdf)

	return &pdf, nil
}

func measureText(draft *gopdf.GoPdf, text string) float64 {
	draft.SetX(0)
	draft.SetY(0)
	draft.Cell(nil, text)

	return draft.GetX()
}

func addPage(pdf *gopdf.GoPdf) {
	pdf.AddPage()

	pdf.SetFillColor(0, 0, 0)
	pdf.RectFromUpperLeftWithStyle(0, 0, pageWidth, pageHeight, "FD")
}

func writeCenteredLine(pdf *gopdf.GoPdf, text string) error {
	pdf.SetFont("default", "", fontSize)
	textWidth, err := pdf.MeasureTextWidth(text)
	if err != nil {
		return err
	}

	pdf.SetX((pageWidth - textWidth) / 2)
	pdf.SetFillColor(255, 255, 255)
	return pdf.Cell(nil, text)
}

func writeCenteredParagraph(pdf *gopdf.GoPdf, text string) error {
	pdf.SetFont("default", "", fontSize)
	lines := strings.Split(text, "\n")
	lines = BreakLongLines(lines, pdf.MeasureTextWidth, contentWidth)
	numLines := len(lines)
	lineHeight := float64(fontSize) * lineSpacing
	paragraphHeight := float64(numLines) * lineHeight
	y0 := (pageHeight - paragraphHeight) / 2

	for index, line := range lines {
		y := y0 + float64(index)*lineHeight
		pdf.SetY(y)
		err := writeCenteredLine(pdf, line)
		if err != nil {
			return err
		}
	}

	return nil
}

func writeHint(pdf *gopdf.GoPdf, text string) error {
	pdf.SetFont("default", "", hintFontSize)

	pdf.SetX(10)
	pdf.SetY(pageHeight - float64(hintFontSize) - 10)
	pdf.SetFillColor(120, 120, 120)
	return pdf.Cell(nil, text)
}

func drawQrCode(pdf *gopdf.GoPdf, content string) {
	qrSize := 400
	var png []byte
	png, err := qrcode.Encode(content, qrcode.Medium, qrSize)
	if err != nil {
		return
	}
	imageHolder, err := gopdf.ImageHolderByBytes(png)
	if err != nil {
		return
	}
	x := (pageWidth - float64(qrSize)) / 2
	y := (pageHeight - float64(qrSize)) / 2
	rect := &gopdf.Rect{W: float64(qrSize), H: float64(qrSize)}
	pdf.ImageByHolder(imageHolder, x, y, rect)
	pdf.SetY(pageHeight - y + (y-float64(fontSize))/2)
	writeCenteredLine(pdf, content)
}

func BuildPDF(textDeck [][]string) (*gopdf.GoPdf, error) {
	pdf, err := createNewPDF()
	if err != nil {
		return nil, err
	}

	for _, song := range textDeck {
		hint := ""
		for _, verse := range song {
			if strings.HasPrefix(verse, "<hint>") && strings.HasSuffix(verse, "</hint>") {
				hint = verse[6 : len(verse)-7]
				continue
			}

			addPage(pdf)
			if hint != "" {
				writeHint(pdf, hint)
				hint = ""
				addPage(pdf)
			}

			isUrl := strings.HasPrefix(verse, "https://") || strings.HasPrefix(verse, "http://")
			if isUrl {
				drawQrCode(pdf, verse)
				continue
			}

			err := writeCenteredParagraph(pdf, verse)
			if err != nil {
				return nil, err
			}
		}
		addPage(pdf)
	}

	return pdf, nil
}
