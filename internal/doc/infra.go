package doc

import (
	"context"
	"fmt"
	"log"
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
	log.Printf("[doc.conv] temp dir: %s", dir)

	inputFile := filepath.Join(dir, "input.docx")
	pdfFile := filepath.Join(dir, "input.pdf")

	// 2. пишем DOC/DOCX
	if err := os.WriteFile(inputFile, data, 0644); err != nil {
		return nil, err
	}
	log.Printf("[doc.conv] wrote DOCX: %s", inputFile)

	// 3. LibreOffice → PDF
	log.Printf("[doc.conv] running libreoffice to PDF...")
	cmd := exec.CommandContext(
		ctx,
		"libreoffice",
		"--headless",
		"--convert-to", "pdf",
		"--outdir", dir,
		inputFile,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("[doc.conv] libreoffice ERROR output=%s", out)
		return nil, fmt.Errorf("libreoffice: %v output=%s", err, out)
	}
	log.Printf("[doc.conv] produced PDF: %s", pdfFile)

	// 4. PDF → JPG через pdftoppm с повышенным DPI
	outBase := filepath.Join(dir, "page")
	log.Printf("[doc.conv] running pdftoppm: pdf=%s base=%s", pdfFile, outBase)

	cmd2 := exec.CommandContext(
		ctx,
		"pdftoppm",
		"-jpeg",
		"-r", "200", // повышаем DPI для читаемости текста
		pdfFile,
		outBase,
	)
	if out, err := cmd2.CombinedOutput(); err != nil {
		log.Printf("[doc.conv] pdftoppm ERROR output=%s", out)
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

		log.Printf("[doc.conv] produced JPEG page=%d file=%s", i, fn)

		pages = append(pages, Page{
			Bytes:    b,
			FileName: fmt.Sprintf("page-%d.jpg", i),
			MimeType: "image/jpeg",
		})
	}

	if len(pages) == 0 {
		return nil, fmt.Errorf("no pages produced from doc")
	}

	log.Printf("[doc.conv] total pages: %d (pdf=%s)", len(pages), pdfFile)

	return pages, nil
}
