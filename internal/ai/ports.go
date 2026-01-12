package ai

import "context"

type Service interface {
	// GetReply получает ответ GPT на новое сообщение пользователя.
	// imageURL — опциональная ссылка на картинку (если nil — речь о тексте/голосе).
	GetReply(ctx context.Context, botID string, telegramID int64, userText string, imageURL *string) (string, error)
	GetReplyWithDirectImage(
		ctx context.Context,
		botID string,
		telegramID int64,
		branch string,
		userText string,
		directImageURL string, // ← ОБЯЗАТЕЛЬНО
	) (string, error)
}
