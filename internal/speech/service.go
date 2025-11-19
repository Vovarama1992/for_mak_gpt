package speech

import (
	"context"
	"fmt"

	"github.com/Vovarama1992/make_ziper/internal/bots"
)

// === Интерфейсы ===

type FromSpeechClient interface {
	Transcribe(ctx context.Context, filePath string) (string, error)
}

type ToSpeechClient interface {
	Synthesize(ctx context.Context, voiceID, text, outPath string) error
}

// === Сервис речи ===

type Service struct {
	stt         FromSpeechClient
	tts         ToSpeechClient
	botsService bots.Service
}

func NewService(stt FromSpeechClient, tts ToSpeechClient, botsSvc bots.Service) *Service {
	return &Service{
		stt:         stt,
		tts:         tts,
		botsService: botsSvc,
	}
}

func (s *Service) Transcribe(ctx context.Context, filePath string) (string, error) {
	return s.stt.Transcribe(ctx, filePath)
}

func (s *Service) Synthesize(ctx context.Context, botID string, text, outPath string) error {
	cfg, err := s.botsService.Get(ctx, botID)
	if err != nil {
		return fmt.Errorf("load bot config: %w", err)
	}
	if cfg == nil {
		return fmt.Errorf("bot config not found for %s", botID)
	}
	if cfg.VoiceID == "" {
		return fmt.Errorf("empty voice_id for %s", botID)
	}

	return s.tts.Synthesize(ctx, cfg.VoiceID, text, outPath)
}
