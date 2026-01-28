package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type DeepgramClient struct {
	apiKey string
	client *http.Client
}

func NewDeepgramClient() *DeepgramClient {
	key := os.Getenv("DEEPGRAM_API_KEY")
	if key == "" {
		panic("DEEPGRAM_API_KEY not set")
	}

	return &DeepgramClient{
		apiKey: key,
		client: &http.Client{},
	}
}

func (c *DeepgramClient) Transcribe(ctx context.Context, filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("read audio file: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.deepgram.com/v1/listen?model=nova-2&smart_format=true&language=ru",
		bytes.NewReader(data),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Token "+c.apiKey)
	req.Header.Set("Content-Type", "audio/ogg")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("deepgram request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("deepgram error: %s", body)
	}

	var parsed struct {
		Results struct {
			Channels []struct {
				Alternatives []struct {
					Transcript string `json:"transcript"`
				} `json:"alternatives"`
			} `json:"channels"`
		} `json:"results"`
	}

	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("decode deepgram: %w", err)
	}

	if len(parsed.Results.Channels) == 0 ||
		len(parsed.Results.Channels[0].Alternatives) == 0 {
		return "", fmt.Errorf("empty transcript")
	}

	return parsed.Results.Channels[0].Alternatives[0].Transcript, nil
}
