package speech

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Перплексити-боту не нужен bot_config: voice_id фиксируем (или берём из ENV).
type PerplexityTTS struct {
	apiKey  string
	voiceID string
	httpCli *http.Client
}

func NewPerplexityTTS() *PerplexityTTS {
	key := os.Getenv("ELEVENLABS_API_KEY")
	if key == "" {
		panic("ELEVENLABS_API_KEY not set")
	}

	voiceID := os.Getenv("PERPLEXITY_VOICE_ID")
	if voiceID == "" {
		voiceID = "EXAVITQu4vr4xnSDxMaL" // Rachel (дефолт)
	}

	return &PerplexityTTS{
		apiKey:  key,
		voiceID: voiceID,
		httpCli: http.DefaultClient,
	}
}

func (t *PerplexityTTS) Synthesize(ctx context.Context, text, outPath string) error {
	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", t.voiceID)
	payload := []byte(fmt.Sprintf(`{"text": %q}`, text))

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	req.Header.Set("xi-api-key", t.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "audio/mpeg")

	resp, err := t.httpCli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("elevenlabs error: %s", string(b))
	}

	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
