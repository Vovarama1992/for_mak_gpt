package doc

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type pythonResp struct {
	Pages []struct {
		FileName string `json:"file_name"`
		Mime     string `json:"mime"`
		Base64   string `json:"base64"`
	} `json:"pages"`
}

type PythonDocConverter struct {
	URL string
}

func NewPythonDocConverter() *PythonDocConverter {
	url := os.Getenv("DOC_SERVICE_URL")
	if url == "" {
		// Фолбек — но не галюна, мы точно знаем имя сервиса из docker-compose
		url = "http://python_doc:8000/convert"
	}
	return &PythonDocConverter{URL: url}
}

func (c *PythonDocConverter) ConvertToImages(
	ctx context.Context,
	data []byte,
) ([]Page, error) {

	log.Printf("[doc.conv] sending %d bytes to %s", len(data), c.URL)

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		c.URL,
		bytes.NewReader(data),
	)
	if err != nil {
		log.Printf("[doc.conv] NEW REQUEST ERROR: %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	// ---- REQUEST START ----
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("[doc.conv] HTTP ERROR: %v", err)
		return nil, fmt.Errorf("python service error: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[doc.conv] python status: %d", resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		log.Printf("[doc.conv] BAD STATUS BODY: %s", string(body))
		return nil, fmt.Errorf("python service bad status %d", resp.StatusCode)
	}

	// ---- JSON PARSE ----
	var out pythonResp
	if err := json.Unmarshal(body, &out); err != nil {
		log.Printf("[doc.conv] JSON ERROR: %v", err)
		return nil, err
	}

	log.Printf("[doc.conv] pages received: %d", len(out.Pages))

	// ---- BASE64 DECODE ----
	var pages []Page
	for i, p := range out.Pages {
		raw, err := decodeBase64(p.Base64)
		if err != nil {
			log.Printf("[doc.conv] BASE64 DECODE ERROR (page %d): %v", i+1, err)
			return nil, err
		}
		log.Printf("[doc.conv] decoded page=%s bytes=%d", p.FileName, len(raw))

		pages = append(pages, Page{
			Bytes:    raw,
			FileName: p.FileName,
			MimeType: p.Mime,
		})
	}

	log.Printf("[doc.conv] DONE total_pages=%d", len(pages))
	return pages, nil
}

func decodeBase64(s string) ([]byte, error) {
	return io.ReadAll(base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(s)))
}
