package ai

import (
	"context"
	"fmt"
	"log"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIClient struct {
	client *openai.Client
}

func NewOpenAIClient() *OpenAIClient {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY not set")
	}
	return &OpenAIClient{
		client: openai.NewClient(apiKey),
	}
}

func (c *OpenAIClient) GetCompletion(ctx context.Context, messages []openai.ChatCompletionMessage) (string, error) {
	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = openai.GPT4oMini // дефолт
	}

	log.Printf("[openai] using model: %s, messages: %d", model, len(messages))

	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    model,
		Messages: messages,
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", nil
	}

	return resp.Choices[0].Message.Content, nil
}

func (c *OpenAIClient) Transcribe(ctx context.Context, filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open audio file: %w", err)
	}
	defer f.Close()

	req := openai.AudioRequest{
		Model:       openai.Whisper1,
		FilePath:    filePath,
		Temperature: 0,
	}

	resp, err := c.client.CreateTranscription(ctx, req)
	if err != nil {
		// Детализированный вывод ошибки
		log.Printf("[whisper] transcription error: %+v", err)
		return "", fmt.Errorf("whisper error: %w", err)
	}

	log.Printf("[whisper] raw response: %#v", resp)
	return resp.Text, nil
}
