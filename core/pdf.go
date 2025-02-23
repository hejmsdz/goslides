package core

import (
	"fmt"
	"strings"

	"github.com/signintech/gopdf"
	"github.com/skip2/go-qrcode"
)

type PageConfig struct {
	PageWidth     float64
	PageHeight    float64
	Margin        float64
	FontSize      int
	HintFontSize  int
	LineSpacing   float64
	Font          string
	VerticalAlign string
}

type ContentSlide struct {
	Type       string `json:"t"`
	ItemIndex  int    `json:"i"`
	VerseIndex int    `json:"v"`
	ChunkIndex int    `json:"c"`
}

type PdfSlides struct {
	pageConfig PageConfig
	goPdf      *gopdf.GoPdf
	lineHeight float64
	maxLines   int
}

func (pdf *PdfSlides) Initialize(pageConfig PageConfig) error {
	pdf.pageConfig = pageConfig
	pdf.goPdf = &gopdf.GoPdf{}

	pdf.lineHeight = float64(pdf.pageConfig.FontSize) * pdf.pageConfig.LineSpacing
	pdf.maxLines = int(pageConfig.PageHeight / pdf.lineHeight)

	pageSize := gopdf.Rect{W: pageConfig.PageWidth, H: pageConfig.PageHeight}

	pdf.goPdf.Start(gopdf.Config{PageSize: pageSize})

	err := pdf.goPdf.AddTTFFont("default", pageConfig.Font)
	if err != nil {
		return err
	}

	pdf.addPage()

	return nil
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

func (pdf *PdfSlides) writeVerse(text string) (int, error) {
	contentWidth := pdf.pageConfig.PageWidth - 2*pdf.pageConfig.Margin
	pdf.goPdf.SetFont("default", "", pdf.pageConfig.FontSize)
	lines := strings.Split(text, "\n")
	lines = BreakLongLines(lines, pdf.goPdf.MeasureTextWidth, contentWidth)

	subPages := SplitLongSlide(lines, pdf.maxLines)

	for i, linesSlice := range subPages {
		if i > 0 {
			pdf.addPage()
		}

		err := pdf.writeAlignedParagraph(linesSlice)
		if err != nil {
			return 0, err
		}
	}

	return len(subPages), nil
}

func (pdf *PdfSlides) writeAlignedParagraph(lines []string) error {
	paragraphHeight := float64(len(lines)) * pdf.lineHeight
	var y0 float64

	switch pdf.pageConfig.VerticalAlign {
	case "top":
		y0 = pdf.pageConfig.Margin
	case "bottom":
		y0 = pdf.pageConfig.PageHeight - paragraphHeight - pdf.pageConfig.Margin
	default:
		y0 = (pdf.pageConfig.PageHeight - paragraphHeight) / 2
	}

	return pdf.writeParagraph(lines, y0)
}

func (pdf *PdfSlides) writeParagraph(lines []string, y0 float64) error {
	offset := float64(pdf.pageConfig.FontSize) * (pdf.pageConfig.LineSpacing - 1) / 2
	for index, line := range lines {
		y := y0 + float64(index)*pdf.lineHeight + offset
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

func BuildPDF(textDeck [][]string, pageConfig PageConfig) (*gopdf.GoPdf, []ContentSlide, error) {
	fmt.Printf("%+v\n", pageConfig)
	pdf := PdfSlides{}
	err := pdf.Initialize(pageConfig)
	if err != nil {
		return nil, nil, err
	}

	contents := make([]ContentSlide, 0)
	contents = append(contents, ContentSlide{Type: "blank", ItemIndex: -1})

	for itemIndex, song := range textDeck {
		hint := ""
		for verseIndex, verse := range song {
			/*
				if strings.HasPrefix(verse, hintStartTag) && strings.HasSuffix(verse, hintEndTag) {
					hint = verse[len(hintStartTag) : len(verse)-len(hintEndTag)]
					continue
				}
			*/

			pdf.addPage()
			if hint != "" {
				pdf.writeHint(hint)
				contents = append(contents, ContentSlide{Type: "hint", ItemIndex: itemIndex})
				hint = ""
				pdf.addPage()
			}

			isUrl := strings.HasPrefix(verse, "https://") || strings.HasPrefix(verse, "http://")
			if isUrl {
				contents = append(contents, ContentSlide{Type: "qr", ItemIndex: itemIndex})
				pdf.drawQrCode(verse)
				continue
			}

			numChunks, err := pdf.writeVerse(verse)
			if err != nil {
				return nil, nil, err
			}

			for i := 0; i < numChunks; i++ {
				contents = append(contents, ContentSlide{
					Type:       "verse",
					ItemIndex:  itemIndex,
					VerseIndex: verseIndex,
					ChunkIndex: i,
				})
			}
		}
		pdf.addPage()
		contents = append(contents, ContentSlide{Type: "blank", ItemIndex: itemIndex})
	}

	return pdf.goPdf, contents, nil
}
