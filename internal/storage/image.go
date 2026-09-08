package storage

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

func SaveImage(file multipart.File, filename string) (string, error) {

	if err := os.MkdirAll("storage", 0755); err != nil {
		return "", err
	}

	path := filepath.Join("storage", filename)

	dst, err := os.Create(path)

	if err != nil {
		return "", err
	}

	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}
	return filename, nil

}
