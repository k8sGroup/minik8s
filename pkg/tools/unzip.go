package tools

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

func Unzip(zipPathName string, destDir string) error {
	zipReader, err := zip.OpenReader(zipPathName)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		fpath := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(fpath, os.ModePerm)
		} else {
			var outFile *os.File
			var inFile io.ReadCloser

			err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm)
			if err != nil {
				goto Err
			}

			inFile, err = f.Open()
			if err != nil {
				goto Err
			}

			outFile, err = os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				goto CloseInFile
			}

			_, err = io.Copy(outFile, inFile)
			if err != nil {
				goto CloseOutFile
			}

		CloseOutFile:
			outFile.Close()
		CloseInFile:
			inFile.Close()
		Err:
			if err != nil {
				return err
			}
		}
	}
	return nil
}
