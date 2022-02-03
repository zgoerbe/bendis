package bendis

import (
	"errors"
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"github.com/zgoerbe/bendis/filesystems"
	"io"
	"net/http"
	"os"
	"path"
)

func (b *Bendis) UploadFile(r *http.Request, destination, field string, fs filesystems.FS) error {
	fileName, err := b.getFileToUpload(r, field)
	if err != nil {
		b.ErrorLog.Println(err)
		return err
	}

	if fs != nil {
		err = fs.Put(fileName, destination)
		if err != nil {
			b.ErrorLog.Println(err)
			return err
		}
	} else {
		err = os.Rename(fileName, fmt.Sprintf("%s/%s", destination, path.Base(fileName)))
		if err != nil {
			b.ErrorLog.Println(err)
			return err
		}
	}

	// delete temp file after upload
	defer func() {
		_ = os.Remove(fileName)
	}()

	return nil
}

func (b *Bendis) getFileToUpload(r *http.Request, fieldName string) (string, error) {
	err := r.ParseMultipartForm(b.config.uploads.maxUploadSize)
	if err != nil {
		return "", err
	}

	file, header, err := r.FormFile(fieldName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// detect the file
	mimeType, err := mimetype.DetectReader(file)
	if err != nil {
		return "", err
	}

	// go back to start of file
	_, err = file.Seek(0, 0)
	if err != nil {
		return "", err
	}

	if !inSlice(b.config.uploads.allowedMimeTypes, mimeType.String()) {
		return "", errors.New("invalid type uploaded")
	}

	dst, err := os.Create(fmt.Sprintf("./tmp/%s", header.Filename))
	if err != nil {
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("./tmp/%s", header.Filename), nil
}

func inSlice(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
