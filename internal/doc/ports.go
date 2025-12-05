package doc

import "context"

type Page struct {
	Bytes    []byte
	FileName string
	MimeType string
}

type Converter interface {
	ConvertToImages(ctx context.Context, data []byte) ([]Page, error)
}
