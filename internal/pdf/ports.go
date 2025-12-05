package pdf

import (
	"context"
	"io"
)

type PDFPage struct {
	Bytes    []byte
	FileName string
	MimeType string
}

type PDFConverter interface {
	ConvertToImages(ctx context.Context, pdf io.Reader) ([]PDFPage, error)
}
