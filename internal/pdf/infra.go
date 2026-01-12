package pdf

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type PopplerPDFConverter struct{}

func NewPopplerPDFConverter() *PopplerPDFConverter {
	return &PopplerPDFConverter{}
}

func (c *PopplerPDFConverter) ConvertToImages(
	ctx context.Context,
	pdf io.Reader,
) ([]PDFPage, error) {

	log.Println("[pdf.conv] reading PDF...")

	buf, err := io.ReadAll(pdf)
	if err != nil {
		log.Printf("[pdf.conv] read ERROR: %v", err)
		return nil, err
	}

	tmpDir, err := os.MkdirTemp("", "pdfconv-*")
	if err != nil {
		log.Printf("[pdf.conv] mktemp ERROR: %v", err)
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	input := filepath.Join(tmpDir, "input.pdf")
	if err := os.WriteFile(input, buf, 0644); err != nil {
		log.Printf("[pdf.conv] write ERROR: %v", err)
		return nil, err
	}

	outBase := filepath.Join(tmpDir, "page")
	log.Printf("[pdf.conv] running pdftoppm input=%s", input)

	cmd := exec.CommandContext(
		ctx,
		"pdftoppm",
		"-r", "120",
		"-jpeg",
		"-jpegopt", "quality=60",
		input,
		outBase,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[pdf.conv] poppler ERROR: %v, output=%s", err, string(out))
		return nil, fmt.Errorf("pdftoppm: %w", err)
	}

	// === FIX: читаем фактические файлы ===

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, "page-") && strings.HasSuffix(name, ".jpg") {
			files = append(files, name)
		}
	}

	sort.Strings(files)

	if len(files) == 0 {
		return nil, fmt.Errorf("no pages generated")
	}

	pages := make([]PDFPage, 0, len(files))
	for i, name := range files {
		b, err := os.ReadFile(filepath.Join(tmpDir, name))
		if err != nil {
			return nil, err
		}

		pages = append(pages, PDFPage{
			Bytes:    b,
			FileName: fmt.Sprintf("page-%d.jpg", i+1),
			MimeType: "image/jpeg",
		})
	}

	return pages, nil
}
