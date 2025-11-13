package speech

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
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
		voice = "21m00Tcm4TlvDq8ikWAM"
	}
	return &ElevenLabsClient{
		apiKey: key,
		voice:  voice,
	}
}

// === SPEECH → TEXT ===
func (c *ElevenLabsClient) Transcribe(ctx context.Context, filePath string) (string, error) {
	url := "https://api.elevenlabs.io/v1/speech-to-text"

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open file fail: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// модель (ОБЯЗАТЕЛЬНО)
	_ = writer.WriteField("model_id", "scribe_v2")

	// файл
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", fmt.Errorf("form file fail: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return "", fmt.Errorf("copy fail: %w", err)
	}

	writer.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return "", fmt.Errorf("request build fail: %w", err)
	}

	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request fail: %w", err)
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("stt failed (%d): %s", resp.StatusCode, string(respBytes))
	}

	var out struct {
		Text string `json:"text"`
	}

	if err := json.Unmarshal(respBytes, &out); err != nil {
		return "", fmt.Errorf("json decode fail: %w raw=%s", err, string(respBytes))
	}

	return out.Text, nil
}

// === TEXT → SPEECH ===
func (c *ElevenLabsClient) Synthesize(ctx context.Context, text, outPath string) error {
	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", c.voice)

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
