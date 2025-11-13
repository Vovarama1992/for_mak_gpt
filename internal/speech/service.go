package speech

import (
	"context"
)

// === Интерфейсы ===

type FromSpeechlient interface {
	Transcribe(ctx context.Context, filePath string) (string, error)
}

type ToSpeechClient interface {
	Synthesize(ctx context.Context, text, outPath string) error
}

// === Единый сервис (и для стт и для ттс) ===

type Service struct {
	stt FromSpeechlient
	tts ToSpeechClient
}

func NewService(stt FromSpeechlient, tts ToSpeechClient) *Service {
	return &Service{
		stt: stt,
		tts: tts,
	}
}

func (s *Service) Transcribe(ctx context.Context, filePath string) (string, error) {
	return s.stt.Transcribe(ctx, filePath)
}

func (s *Service) Synthesize(ctx context.Context, text, outPath string) error {
	return s.tts.Synthesize(ctx, text, outPath)
}
