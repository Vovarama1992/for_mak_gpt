package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type PerplexityClient struct {
	apiKey string
	client *http.Client
}

func NewPerplexityClient() *PerplexityClient {
	key := os.Getenv("PERPLEXITY_API_KEY")
	if key == "" {
		panic("PERPLEXITY_API_KEY not set")
	}

	return &PerplexityClient{
		apiKey: key,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

type perplexityRequest struct {
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
}

type perplexityResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (c *PerplexityClient) Ask(ctx context.Context, question string) (string, error) {
	reqBody := perplexityRequest{
		Model: "sonar",
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{
				Role:    "user",
				Content: question,
			},
		},
	}

	b, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.perplexity.ai/chat/completions",
		bytes.NewBuffer(b),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("perplexity status: %s", resp.Status)
	}

	var out perplexityResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}

	if len(out.Choices) == 0 {
		return "", nil
	}

	return out.Choices[0].Message.Content, nil
}
