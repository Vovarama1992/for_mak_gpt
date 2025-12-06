package doc

import "context"

type Service struct {
	conv Converter
}

func NewService(conv Converter) *Service {
	return &Service{conv: conv}
}

func (s *Service) Convert(ctx context.Context, data []byte) (string, error) {
	return s.conv.ConvertToText(ctx, data)
}
