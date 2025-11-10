package ai

import "context"

type Service interface {
    // GetReply получает ответ от GPT на новое сообщение пользователя.
    // При этом сервис сам подтягивает историю из БД и сохраняет новый ответ.
    GetReply(ctx context.Context, botID string, telegramID int64, userText string) (string, error)
}