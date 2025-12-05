package pdf

import (
	"context"
	"fmt"
	"io"
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

	// 1. читаем PDF в память
	buf, err := io.ReadAll(pdf)
	if err != nil {
		return nil, err
	}

	// 2. создаём уникальный temp-dir
	tmpDir, err := os.MkdirTemp("", "pdfconv-*")
	if err != nil {
		return nil, err
	}
	// подчистим потом
	defer os.RemoveAll(tmpDir)

	// путь ввода
	input := filepath.Join(tmpDir, "input.pdf")
	if err := os.WriteFile(input, buf, 0644); err != nil {
		return nil, err
	}

	// базовый путь вывода
	outBase := filepath.Join(tmpDir, "page")

	// 3. запускаем poppler
	cmd := exec.CommandContext(
		ctx,
		"pdftoppm",
		input,
		outBase,
		"-jpeg",
	)

	if err := cmd.Run(); err != nil {
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
