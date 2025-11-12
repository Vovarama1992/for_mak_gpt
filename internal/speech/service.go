package speech

import "context"

type Service struct {
	client Client
}

func NewService(client Client) *Service {
	return &Service{client: client}
}

func (s *Service) Transcribe(ctx context.Context, filePath string) (string, error) {
	return s.client.Transcribe(ctx, filePath)
}

func (s *Service) Synthesize(ctx context.Context, text, outPath string) error {
	return s.client.Synthesize(ctx, text, outPath)
}