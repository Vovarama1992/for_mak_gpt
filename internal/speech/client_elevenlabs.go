package speech

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type ElevenLabsClient struct {
	apiKey string
}

func NewElevenLabsClient() *ElevenLabsClient {
	key := os.Getenv("ELEVENLABS_API_KEY")
	if key == "" {
		panic("ELEVENLABS_API_KEY not set")
	}

	return &ElevenLabsClient{
		apiKey: key,
	}
}

// TEXT → SPEECH
func (c *ElevenLabsClient) Synthesize(ctx context.Context, voiceID, text, outPath string) error {
	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voiceID)

	payload := []byte(fmt.Sprintf(`{"text": %q}`, text))

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "audio/mpeg")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("tts failed: %s", string(b))
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return err
	}

	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
