package compress

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func walk(projectDir string, writer *tar.Writer) filepath.WalkFunc {
	return func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create a header for the file or directory
		header, err := tar.FileInfoHeader(info, path)
		if err != nil {
			return err
		}

		// Adjust the header name to be relative to the project directory
		header.Name, _ = filepath.Rel(projectDir, path)

		// Write the header
		if err := writer.WriteHeader(header); err != nil {
			return err
		}

		// If it's a file, write its contents
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(writer, file)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func CompressDirectory(projectDir string) (*bytes.Buffer, error) {
	fmt.Println(projectDir)
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// Walk the project directory and add files to the tar writer
	err := filepath.Walk(projectDir, walk(projectDir, tw))
	if err != nil {
		return nil, fmt.Errorf("")

	}
	return &buf, nil
}
