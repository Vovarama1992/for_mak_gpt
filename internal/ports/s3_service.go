package ports

import (
	"context"
	"io"
)

type S3Service interface {
	// хранение пользовательских изображений (фото, pdf, word и т.д.)
	SaveImage(
		ctx context.Context,
		botID string,
		telegramID int64,
		file io.Reader,
		filename string,
		contentType string,
	) (string, error)

	// хранение ассетов бота (welcome video, картинки и т.п.)
	SaveBotAsset(
		ctx context.Context,
		botID string,
		file io.Reader,
		filename string,
	) (string, error)
}
