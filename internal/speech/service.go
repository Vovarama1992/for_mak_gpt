package speech

import (
	"context"
)

// === Интерфейсы ===

type STTClient interface {
	Transcribe(ctx context.Context, filePath string) (string, error)
}

type TTSClient interface {
	Synthesize(ctx context.Context, text, outPath string) error
}

// === Единый сервис (и для стт и для ттс) ===

type Service struct {
	stt STTClient
	tts TTSClient
}

func NewService(stt STTClient, tts TTSClient) *Service {
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
