package speech

import (
	"context"
	"fmt"

	"github.com/Vovarama1992/make_ziper/internal/bots"
	"github.com/Vovarama1992/make_ziper/internal/error_notificator"
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
	Notifier    error_notificator.Notificator
}

func NewService(
	stt FromSpeechClient,
	tts ToSpeechClient,
	botsSvc bots.Service,
	notifier error_notificator.Notificator,
) *Service {
	return &Service{
		stt:         stt,
		tts:         tts,
		botsService: botsSvc,
		Notifier:    notifier,
	}
}

func (s *Service) Transcribe(ctx context.Context, botID string, filePath string) (string, error) {
	result, err := s.stt.Transcribe(ctx, filePath)
	if err != nil {
		s.Notifier.Notify(ctx, botID, err, "Ошибка при транскрипции аудио (ASR)")
		return "", err
	}
	return result, nil
}

func (s *Service) Synthesize(ctx context.Context, botID string, text, outPath string) error {
	cfg, err := s.botsService.Get(ctx, botID)
	if err != nil {
		s.Notifier.Notify(ctx, botID, err, "Ошибка загрузки bot_config в Synthesize")
		return fmt.Errorf("load bot config: %w", err)
	}
	if cfg == nil {
		err := fmt.Errorf("bot config not found for %s", botID)
		s.Notifier.Notify(ctx, botID, err, "bot config отсутствует")
		return err
	}
	if cfg.VoiceID == "" {
		err := fmt.Errorf("empty voice_id for %s", botID)
		s.Notifier.Notify(ctx, botID, err, "voice_id пустой — синтез невозможен")
		return err
	}

	err = s.tts.Synthesize(ctx, cfg.VoiceID, text, outPath)
	if err != nil {
		s.Notifier.Notify(ctx, botID, err, "Ошибка синтеза речи (TTS)")
		return err
	}

	return nil
}
