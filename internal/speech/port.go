package speech

import "context"

type Client interface {
	Transcribe(ctx context.Context, filePath string) (string, error) // голос → текст
	Synthesize(ctx context.Context, text, outPath string) error     // текст → голос (сохраняет файл)
}