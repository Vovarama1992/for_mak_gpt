package doc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type pythonResp struct {
	Text string `json:"text"`
}

type PythonDocConverter struct {
	URL string
}

func NewPythonDocConverter() *PythonDocConverter {
	url := os.Getenv("DOC_SERVICE_URL")
	if url == "" {
		// Фолбек (строго НЕ галюн)
		url = "http://python_doc:8000/convert"
	}
	return &PythonDocConverter{URL: url}
}

func (c *PythonDocConverter) ConvertToText(
	ctx context.Context,
	data []byte,
) (string, error) {

	log.Printf("[doc.conv] sending %d bytes to %s", len(data), c.URL)

	req, err := http.NewRequestWithContext(ctx, "POST", c.URL, bytes.NewReader(data))
	if err != nil {
		log.Printf("[doc.conv] NEW REQUEST ERROR: %v", err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	// ---- HTTP ----
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("[doc.conv] HTTP ERROR: %v", err)
		return "", fmt.Errorf("python service error: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[doc.conv] python status: %d", resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		log.Printf("[doc.conv] BAD STATUS BODY: %s", string(body))
		return "", fmt.Errorf("python bad status %d", resp.StatusCode)
	}

	// ---- PARSE JSON ----
	var out pythonResp
	if err := json.Unmarshal(body, &out); err != nil {
		log.Printf("[doc.conv] JSON ERROR: %v", err)
		return "", err
	}

	log.Printf("[doc.conv] text length received: %d", len(out.Text))

	return out.Text, nil
}
