package internal

import (
	"github.com/farishadibrata/maroto/pkg/color"
	"github.com/farishadibrata/maroto/pkg/consts"
	"github.com/farishadibrata/maroto/pkg/props"
)

const (
	lineHeight = 1.0
)

// MarotoGridPart is the abstraction to deal with the gris system inside the table list.
type MarotoGridPart interface {
	// Grid System.
	Row(height float64, closure func())
	Col(width uint, closure func())
	ColSpace(width uint)

	// Helpers.
	SetBackgroundColor(color color.Color)
	GetCurrentOffset() float64
	GetPageSize() (width float64, height float64)
	GetPageMargins() (left float64, top float64, right float64, bottom float64)

	// Outside Col/Row Components.
	Line(spaceHeight float64, line ...props.Line)

	// Inside Col/Row Components.
	Text(text string, prop ...props.Text)
	GetCurrentPage() int
}

// TableList is the abstraction to create a table with header and contents.
type TableList interface {
	Create(header []string, contents [][]string, defaultFontFamily string, prop ...props.TableList)
	BindGrid(part MarotoGridPart)
}

type tableList struct {
	pdf           MarotoGridPart
	text          Text
	font          Font
	lastPageIndex int
}

// NewTableList create a TableList.
func NewTableList(text Text, font Font) *tableList {
	return &tableList{
		text: text,
		font: font,
	}
}

// BindGrid bind the grid system to TableList.
func (s *tableList) BindGrid(pdf MarotoGridPart) {
	s.pdf = pdf
}

// Create method creates a header section with a list of strings and
// create many rows with contents.
func (s *tableList) Create(header []string, contents [][]string, defaultFontFamily string, prop ...props.TableList) {
	if len(header) == 0 {
		return
	}

	if len(contents) == 0 {
		return
	}

	tableProp := props.TableList{}

	if len(prop) > 0 {
		tableProp = prop[0]
	}
	if tableProp.ShowHeaderOnNewPage != nil {
		s.lastPageIndex = s.pdf.GetCurrentPage()
	}

	tableProp.MakeValid(header, defaultFontFamily)
	headerHeight := s.calcLinesHeight(header, tableProp.HeaderProp, tableProp.Align)
	if tableProp.HeaderHeight != nil {
		headerHeight = headerHeight + float64(*tableProp.HeaderHeight)
	}
	// Draw header.
	if !tableProp.DisableHeader {
		s.pdf.Row(headerHeight+0.5, func() {
			for i, h := range header {
				hs := h

				if tableProp.HeaderBackground != nil {
					s.pdf.SetBackgroundColor(*tableProp.HeaderBackground)
				}

				s.pdf.Col(tableProp.HeaderProp.GridSizes[i], func() {
					reason := hs
					s.pdf.Text(reason, tableProp.HeaderProp.ToTextProp(tableProp.Align, 0, false, 0.0, tableProp.Left))
				})
				s.pdf.SetBackgroundColor(color.NewWhite())
			}
		})
	}

	// Define space between header and contents.
	s.pdf.Row(tableProp.HeaderContentSpace, func() {
		s.pdf.ColSpace(0)
	})

	// Draw contents.
	if tableProp.DisableContent {
		return
	}
	for index, content := range contents {
		if tableProp.ShowHeaderOnNewPage != nil {
			currentPage := s.pdf.GetCurrentPage()
			if s.lastPageIndex != currentPage {
				// Draw header
				tableProp.ShowHeaderOnNewPage()
				// s.pdf.Row(headerHeight+1, func() {
				// 	for i, h := range header {
				// 		hs := h

				// 		s.pdf.Col(tableProp.HeaderProp.GridSizes[i], func() {
				// 			reason := hs
				// 			s.pdf.Text(reason, tableProp.HeaderProp.ToTextProp(tableProp.Align, 0, false, 0.0, tableProp.Left))
				// 		})
				// 	}
				// })
				s.lastPageIndex = currentPage
			}
		}

		contentHeight := s.calcLinesHeight(content, tableProp.ContentProp, tableProp.Align)
		contentHeightPadded := contentHeight + tableProp.VerticalContentPadding

		if tableProp.AlternatedBackground != nil && index%2 == 0 {
			s.pdf.SetBackgroundColor(*tableProp.AlternatedBackground)
		}

		s.pdf.Row(contentHeightPadded+1, func() {
			for i, c := range content {
				cs := c

				s.pdf.Col(tableProp.ContentProp.GridSizes[i], func() {
					s.pdf.Text(cs, tableProp.ContentProp.ToTextProp(tableProp.Align, tableProp.VerticalContentPadding/2.0, false, 0.0, tableProp.Left))
				})
			}
		})

		if tableProp.AlternatedBackground != nil && index%2 == 0 {
			s.pdf.SetBackgroundColor(color.NewWhite())
		}

		if tableProp.Line {
			s.pdf.Line(lineHeight, tableProp.LineProp)
		}
	}
}

func (s *tableList) calcLinesHeight(textList []string, contentProp props.TableListContent, align consts.Align) float64 {
	maxLines := 1.0

	left, _, right, _ := s.pdf.GetPageMargins()
	width, _ := s.pdf.GetPageSize()
	usefulWidth := width - left - right

	textProp := contentProp.ToTextProp(align, 0, false, 0.0, 0)

	for i, text := range textList {
		gridSize := float64(contentProp.GridSizes[i])
		percentSize := gridSize / consts.MaxGridSum
		colWidth := usefulWidth * percentSize
		qtdLines := float64(s.text.GetLinesQuantity(text, textProp, colWidth))
		if qtdLines > maxLines {
			maxLines = qtdLines
		}
		// Special Rule for overlap 4.2
		if contentProp.Spacing != 0 {
			maxLines = qtdLines + float64(contentProp.Spacing)
		}
	}

	_, _, fontSize := s.font.GetFont()

	// Font size corrected by the scale factor from "mm" inside gofpdf f.k.
	fontHeight := fontSize / s.font.GetScaleFactor()

	return fontHeight * maxLines
}
