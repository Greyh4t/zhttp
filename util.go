package zhttp

import (
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var quoteEscaper = strings.NewReplacer(`\`, `\\`, `"`, `\"`)

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

// FileFromDisk read file from disk and detect mime with filename
func FileFromDisk(filePath string) (*File, error) {
	filePath = filepath.Clean(filePath)
	fd, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	file := &File{
		Name:     fd.Name(),
		Contents: fd,
	}
	file.Mime = mime.TypeByExtension(filepath.Ext(file.Name))

	return file, nil
}

// MustProxy convert a url string to *url.URL, if there has an error, will panic
func MustProxy(rawURL string) *url.URL {
	urlObj, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return urlObj
}
