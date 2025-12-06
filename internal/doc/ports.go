package doc

import "context"

type Converter interface {
	ConvertToText(ctx context.Context, data []byte) (string, error)
}

type DocConverter interface {
	Convert(ctx context.Context, data []byte) (string, error)
}
