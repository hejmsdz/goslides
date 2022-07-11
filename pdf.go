package main

import (
	"fmt"
	"strings"

	"github.com/signintech/gopdf"
	"github.com/skip2/go-qrcode"
)

type PageConfig struct {
	PageWidth    float64
	PageHeight   float64
	Margin       float64
	FontSize     int
	HintFontSize int
	LineSpacing  float64
	Font         string
}

type PdfSlides struct {
	pageConfig PageConfig
	goPdf      *gopdf.GoPdf
}

func (pdf *PdfSlides) Initialize(pageConfig PageConfig) error {
	pdf.pageConfig = pageConfig
	pdf.goPdf = &gopdf.GoPdf{}

	pageSize := gopdf.Rect{W: pageConfig.PageWidth, H: pageConfig.PageHeight}

	pdf.goPdf.Start(gopdf.Config{PageSize: pageSize})

	err := pdf.goPdf.AddTTFFont("default", pageConfig.Font)
	if err != nil {
		return err
	}

	pdf.addPage()

	return nil
}

func measureText(draft *gopdf.GoPdf, text string) float64 {
	draft.SetX(0)
	draft.SetY(0)
	draft.Cell(nil, text)

	return draft.GetX()
}

func (pdf *PdfSlides) addPage() {
	pdf.goPdf.AddPage()

	pdf.goPdf.SetFillColor(0, 0, 0)
	pdf.goPdf.RectFromUpperLeftWithStyle(0, 0, pdf.pageConfig.PageWidth, pdf.pageConfig.PageHeight, "FD")
}

func (pdf *PdfSlides) writeCenteredLine(text string) error {
	pdf.goPdf.SetFont("default", "", pdf.pageConfig.FontSize)
	textWidth, err := pdf.goPdf.MeasureTextWidth(text)
	if err != nil {
		return err
	}

	pdf.goPdf.SetX((pdf.pageConfig.PageWidth - textWidth) / 2)
	pdf.goPdf.SetFillColor(255, 255, 255)
	return pdf.goPdf.Cell(nil, text)
}

func (pdf *PdfSlides) writeCenteredParagraph(text string) error {
	contentWidth := pdf.pageConfig.PageWidth - 2*pdf.pageConfig.Margin
	pdf.goPdf.SetFont("default", "", pdf.pageConfig.FontSize)
	lines := strings.Split(text, "\n")
	lines = BreakLongLines(lines, pdf.goPdf.MeasureTextWidth, contentWidth)
	numLines := len(lines)
	lineHeight := float64(pdf.pageConfig.FontSize) * pdf.pageConfig.LineSpacing
	paragraphHeight := float64(numLines) * lineHeight
	y0 := (pdf.pageConfig.PageHeight - paragraphHeight) / 2

	for index, line := range lines {
		y := y0 + float64(index)*lineHeight
		pdf.goPdf.SetY(y)
		err := pdf.writeCenteredLine(line)
		if err != nil {
			return err
		}
	}

	return nil
}

func (pdf *PdfSlides) writeHint(text string) error {
	pdf.goPdf.SetFont("default", "", pdf.pageConfig.HintFontSize)

	pdf.goPdf.SetX(10)
	pdf.goPdf.SetY(pdf.pageConfig.PageHeight - float64(pdf.pageConfig.HintFontSize) - 10)
	pdf.goPdf.SetFillColor(120, 120, 120)
	return pdf.goPdf.Cell(nil, text)
}

func (pdf *PdfSlides) drawQrCode(content string) {
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
	x := (pdf.pageConfig.PageWidth - float64(qrSize)) / 2
	y := (pdf.pageConfig.PageHeight - float64(qrSize)) / 2
	rect := &gopdf.Rect{W: float64(qrSize), H: float64(qrSize)}
	pdf.goPdf.ImageByHolder(imageHolder, x, y, rect)
	pdf.goPdf.SetY(pdf.pageConfig.PageHeight - y + (y-float64(pdf.pageConfig.FontSize))/2)
	pdf.writeCenteredLine(content)
}

func BuildPDF(textDeck [][]string, pageConfig PageConfig) (*gopdf.GoPdf, error) {
	pdf := PdfSlides{}
	err := pdf.Initialize(pageConfig)
	fmt.Printf("%+v", pdf)
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

			pdf.addPage()
			if hint != "" {
				pdf.writeHint(hint)
				hint = ""
				pdf.addPage()
			}

			isUrl := strings.HasPrefix(verse, "https://") || strings.HasPrefix(verse, "http://")
			if isUrl {
				pdf.drawQrCode(verse)
				continue
			}

			err := pdf.writeCenteredParagraph(verse)
			if err != nil {
				return nil, err
			}
		}
		pdf.addPage()
	}

	return pdf.goPdf, nil
}
