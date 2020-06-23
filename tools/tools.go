package tools

import (
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/greyh4t/zhttp"
)

// DeepGetCookie check all responses in the redirect and return the first matching url and cookie
func DeepGetCookie(resp *http.Response, name string) (string, string, bool) {
	if resp == nil {
		return "", "", false
	}

	req := resp.Request
	cookies := resp.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == name {
			return req.URL.String(), cookie.Value, true
		}
	}

	return DeepGetCookie(req.Response, name)
}

// DeepGetHeader check all responses in the redirect and return the first matching url and header
func DeepGetHeader(resp *http.Response, key string) (string, string, bool) {
	if resp == nil {
		return "", "", false
	}

	req := resp.Request
	values, ok := resp.Header[http.CanonicalHeaderKey(key)]
	if ok {
		return req.URL.String(), strings.Join(values, ", "), true
	}

	return DeepGetHeader(req.Response, key)
}

// CookiesFromRaw parse a cookie in string format to []*http.Cookie
func CookiesFromRaw(rawCookie string, domain string) []*http.Cookie {
	list := strings.Split(rawCookie, ";")
	if len(list) == 0 {
		return nil
	}

	cookies := make([]*http.Cookie, len(list))
	for i, item := range list {
		pairs := strings.SplitN(strings.TrimSpace(item), "=", 2)
		cookie := &http.Cookie{
			Name:   pairs[0],
			Domain: domain,
		}

		if len(pairs) == 2 {
			cookie.Value = pairs[1]
		}

		cookies[i] = cookie
	}

	return cookies
}

// FileFromDisk read file from disk and detect mime with filename
func FileFromDisk(filePath string) (*zhttp.File, error) {
	filePath = filepath.Clean(filePath)
	fd, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	file := &zhttp.File{
		Name:     fd.Name(),
		Contents: fd,
	}
	file.Mime = mime.TypeByExtension(filepath.Ext(file.Name))

	return file, nil
}

// SaveToFile save reader's content to file
func SaveToFile(r io.Reader, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err := io.Copy(f, r); err != nil && err != io.EOF {
		return err
	}

	return nil
}
