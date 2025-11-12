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
	voice  string
}

func NewElevenLabsClient() *ElevenLabsClient {
	key := os.Getenv("ELEVENLABS_API_KEY")
	if key == "" {
		panic("ELEVENLABS_API_KEY not set")
	}
	voice := os.Getenv("ELEVENLABS_VOICE_ID")
	if voice == "" {
		voice = "21m00Tcm4TlvDq8ikWAM" // Rachel (дефолт)
	}
	return &ElevenLabsClient{
		apiKey: key,
		voice:  voice,
	}
}

// Transcribe — пока заглушка (можно потом подключить Whisper или другой STT).
func (c *ElevenLabsClient) Transcribe(ctx context.Context, filePath string) (string, error) {
	return "", fmt.Errorf("transcribe not implemented for ElevenLabs")
}

// Synthesize — превращает текст в голосовой файл (mp3 или wav)
func (c *ElevenLabsClient) Synthesize(ctx context.Context, text, outPath string) error {
	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", c.voice)

	payload := []byte(fmt.Sprintf(`{"text": %q}`, text))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
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
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("tts failed: %s", string(body))
	}

	dir := filepath.Dir(outPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	return err
}