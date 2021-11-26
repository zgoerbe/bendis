package bendis

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"path"
	"path/filepath"
)

func (b *Bendis) WriteJSON(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_, err = w.Write(out)
	if err != nil {
		return err
	}

	return nil
}

func (b *Bendis) WriteXML(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := xml.MarshalIndent(data, "", "     ")
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)

	_, err = w.Write(out)
	if err != nil {
		return err
	}

	return nil
}

func (b *Bendis) DownloadFile(w http.ResponseWriter, r *http.Request, pathToFile, fileName string) error {
	fp := path.Join(pathToFile, fileName)
	fileToServe := filepath.Clean(fp)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; file=\"%s\"", fileName))
	http.ServeFile(w, r, fileToServe)
	return nil
}

func (b *Bendis) Error404(w http.ResponseWriter, r *http.Request) {
	b.ErrorStatus(w, http.StatusNotFound)
}

func (b *Bendis) Error500(w http.ResponseWriter, r *http.Request) {
	b.ErrorStatus(w, http.StatusInternalServerError)
}

func (b *Bendis) ErrorUnauthorized(w http.ResponseWriter, r *http.Request) {
	b.ErrorStatus(w, http.StatusUnauthorized)
}

func (b *Bendis) ErrorForbidden(w http.ResponseWriter, r *http.Request) {
	b.ErrorStatus(w, http.StatusForbidden)
}

func (b *Bendis) ErrorStatus(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
