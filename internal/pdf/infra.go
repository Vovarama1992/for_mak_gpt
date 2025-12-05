package pdf

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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

	cmd := exec.CommandContext(ctx, "pdftoppm", input, outBase, "-jpeg")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[pdf.conv] poppler ERROR: %v, output=%s", err, string(out))
		return nil, fmt.Errorf("pdftoppm: %w", err)
	}

	// 4. собираем page-1.jpg, page-2.jpg, ...
	var pages []PDFPage
	for i := 1; ; i++ {
		fn := fmt.Sprintf("%s-%d.jpg", outBase, i)

		b, err := os.ReadFile(fn)
		if err != nil {
			break // страниц больше нет
		}

		pages = append(pages, PDFPage{
			Bytes:    b,
			FileName: fmt.Sprintf("page-%d.jpg", i),
			MimeType: "image/jpeg",
		})
	}

	if len(pages) == 0 {
		return nil, fmt.Errorf("no pages generated")
	}

	return pages, nil
}
