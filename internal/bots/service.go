package bots

import (
	"context"
	"io"

	"github.com/Vovarama1992/make_ziper/internal/ports"
)

type service struct {
	repo Repo
	s3   ports.S3Service
}

func NewService(repo Repo, s3 ports.S3Service) Service {
	return &service{
		repo: repo,
		s3:   s3,
	}
}

func (s *service) Create(ctx context.Context, in *CreateInput) (*BotConfig, error) {
	return s.repo.Create(ctx, in)
}

func (s *service) ListAll(ctx context.Context) ([]*BotConfig, error) {
	return s.repo.ListAll(ctx)
}

func (s *service) Get(ctx context.Context, botID string) (*BotConfig, error) {
	return s.repo.Get(ctx, botID)
}

func (s *service) Update(ctx context.Context, in *UpdateInput) (*BotConfig, error) {
	return s.repo.Update(ctx, in)
}

// =========================================================
// UploadWelcomeVideo — загрузка файла в S3 +
// обновление bot_configs.welcome_video_url
// =========================================================
func (s *service) UploadWelcomeVideo(
	ctx context.Context,
	botID string,
	file io.Reader,
	filename string,
) (string, error) {

	// 1) Загружаем в S3
	url, err := s.s3.SaveBotAsset(ctx, botID, file, filename)
	if err != nil {
		return "", err
	}

	// 2) Обновляем запись бота
	in := &UpdateInput{
		BotID:        botID,
		WelcomeVideo: &url,
	}

	if _, err := s.repo.Update(ctx, in); err != nil {
		return "", err
	}

	return url, nil
}
