package doc

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type LibreOfficeConverter struct{}

func NewLibreOfficeConverter() *LibreOfficeConverter {
	return &LibreOfficeConverter{}
}

func (c *LibreOfficeConverter) ConvertToImages(
	ctx context.Context,
	data []byte,
) ([]Page, error) {

	// 1. временная директория
	dir, err := os.MkdirTemp("", "docconv-*")
	if err != nil {
		return nil, err
	}

	inputFile := filepath.Join(dir, "input.docx")
	pdfFile := filepath.Join(dir, "input.pdf")

	// 2. пишем DOC/DOCX
	if err := os.WriteFile(inputFile, data, 0644); err != nil {
		return nil, err
	}

	// 3. LibreOffice → PDF
	cmd := exec.CommandContext(
		ctx,
		"libreoffice",
		"--headless",
		"--convert-to", "pdf",
		"--outdir", dir,
		inputFile,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("libreoffice: %v output=%s", err, out)
	}

	// 4. PDF → JPG через pdftoppm
	cmd2 := exec.CommandContext(
		ctx,
		"pdftoppm",
		pdfFile,
		filepath.Join(dir, "page"),
		"-jpeg",
	)
	if out, err := cmd2.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("pdftoppm: %v output=%s", err, out)
	}

	// 5. собираем страницы
	var pages []Page
	for i := 1; ; i++ {
		fn := filepath.Join(dir, fmt.Sprintf("page-%d.jpg", i))
		b, err := os.ReadFile(fn)
		if err != nil {
			break
		}
		pages = append(pages, Page{
			Bytes:    b,
			FileName: fmt.Sprintf("page-%d.jpg", i),
			MimeType: "image/jpeg",
		})
	}

	if len(pages) == 0 {
		return nil, fmt.Errorf("no pages produced from doc")
	}

	return pages, nil
}
