package extractor

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	securejoin "github.com/cyphar/filepath-securejoin"
)

type zipExtractor struct{}

func NewZip() Extractor {
	return &zipExtractor{}
}

func (e *zipExtractor) Extract(src, dest string) error {
	srcType, err := mimeType(src)
	if err != nil {
		return err
	}

	switch srcType {
	case "application/zip":
		err := extractZip(src, dest)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("%s is not a zip archive: %s", src, srcType)
	}

	return nil
}

func extractZip(src, dest string) error {
	files, err := zip.OpenReader(src)
	if err != nil {
		return err
	}

	defer files.Close()

	for _, file := range files.File {
		err = func() error {
			readCloser, err := file.Open()
			if err != nil {
				return err
			}
			defer readCloser.Close()

			return extractZipArchiveFile(file, dest, readCloser)
		}()

		if err != nil {
			return err
		}
	}

	return nil
}

func extractZipArchiveFile(file *zip.File, dest string, input io.Reader) error {
	filePath, err := securejoin.SecureJoin(dest, file.Name)
	if err != nil {
		return err
	}
	fileInfo := file.FileInfo()

	if fileInfo.IsDir() {
		err = os.MkdirAll(filePath, fileInfo.Mode())
		if err != nil {
			return err
		}
	} else {
		err = os.MkdirAll(filepath.Dir(filePath), 0755)
		if err != nil {
			return err
		}

		if fileInfo.Mode()&os.ModeSymlink != 0 {
			linkName, err := io.ReadAll(input)
			if err != nil {
				return err
			}
			return os.Symlink(string(linkName), filePath)
		}

		fileCopy, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileInfo.Mode())
		if err != nil {
			return err
		}
		defer fileCopy.Close()

		_, err = io.Copy(fileCopy, input)
		if err != nil {
			return err
		}
	}

	return nil
}
