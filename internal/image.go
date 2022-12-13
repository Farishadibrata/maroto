package internal

import (
	"bytes"
	"encoding/base64"
	"errors"
	"math/rand"

	"github.com/farishadibrata/maroto/internal/fpdf"
	"github.com/farishadibrata/maroto/pkg/consts"
	"github.com/farishadibrata/maroto/pkg/props"
	"github.com/jung-kurt/gofpdf"
)

// Image is the abstraction which deals of how to add images in a PDF.
type Image interface {
	AddFromFile(path string, cell Cell, prop props.Rect) (err error)
	AddFromBase64(stringBase64 string, cell Cell, prop props.Rect, extension consts.Extension) (err error)
}

type image struct {
	pdf  fpdf.Fpdf
	math Math
}

// NewImage create an Image.
func NewImage(pdf fpdf.Fpdf, math Math) *image {
	return &image{
		pdf,
		math,
	}
}

const letterBytes = "ASD"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// AddFromFile open an image from disk and add to PDF.
func (s *image) AddFromFile(path string, cell Cell, prop props.Rect) error {
	info := s.pdf.RegisterImageOptions(path, gofpdf.ImageOptions{
		ReadDpi:   false,
		ImageType: "",
	})

	if info == nil {
		return errors.New("could not register image options, maybe path/name is wrong")
	}

	s.addImageToPdf(path, info, cell, prop)
	return nil
}

// AddFromBase64 use a base64 string to add to PDF.
func (s *image) AddFromBase64(stringBase64 string, cell Cell, prop props.Rect, extension consts.Extension) error {
	//WARNING: Had to change from uuid to string because it caused an error in pdfcpu, for weird reason, i cant use random string.
	imageID := "PLEASEREPLACEMEWHENYOUFIXEDIT"
	ss, _ := base64.StdEncoding.DecodeString(stringBase64)

	info := s.pdf.RegisterImageOptionsReader(
		imageID,
		gofpdf.ImageOptions{
			ReadDpi:   false,
			ImageType: string(extension),
		},
		bytes.NewReader(ss),
	)

	if info == nil {
		return errors.New("could not register image options, maybe path/name is wrong")
	}

	s.addImageToPdf(imageID, info, cell, prop)
	return nil
}

func (s *image) addImageToPdf(imageLabel string, info *gofpdf.ImageInfoType, cell Cell, prop props.Rect) {
	var x, y, w, h float64
	if prop.Center {
		x, y, w, h = s.math.GetRectCenterColProperties(info.Width(), info.Height(), cell.Width, cell.Height, cell.X, prop.Percent)
	} else {
		x, y, w, h = s.math.GetRectNonCenterColProperties(info.Width(), info.Height(), cell.Width, cell.Height, cell.X, prop)
	}
	s.pdf.Image(imageLabel, x, y+cell.Y, w, h, false, "", 0, "")
}
