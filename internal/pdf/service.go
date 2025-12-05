package pdf

import (
	"context"
	"io"
)

type PDFService struct {
	conv PDFConverter
}

func NewPDFService(c PDFConverter) *PDFService {
	return &PDFService{conv: c}
}

func (s *PDFService) Convert(ctx context.Context, pdf io.Reader) ([]PDFPage, error) {
	return s.conv.ConvertToImages(ctx, pdf)
}
