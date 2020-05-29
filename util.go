package zhttp

import (
	"io"
	"mime"
	"net/http"
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

// MustProxy convert scheme and url string to P.
// If there have any error, will panic
func MustProxy(proxies M) P {
	if len(proxies) > 0 {
		proxiesMap := P{}
		for scheme, proxyURL := range proxies {
			urlObj, err := url.Parse(proxyURL)
			if err != nil {
				panic(err)
			}
			proxiesMap[scheme] = urlObj
		}
		return proxiesMap
	}
	return nil
}

// RawHTTPRequest format the http.Request to string.
// Notice, the order of headers is not strictly consistent
func RawHTTPRequest(req *http.Request) string {
	var buf strings.Builder
	buf.WriteString(req.Method + " " + req.URL.RequestURI() + " " + req.Proto + "\r\n")

	if req.Host != "" {
		buf.WriteString("Host: " + req.Host + "\r\n")
	} else {
		buf.WriteString("Host: " + req.URL.Host + "\r\n")
	}

	req.Header.Write(&buf)
	buf.WriteString("\r\n")

	if req.GetBody != nil {
		rc, err := req.GetBody()
		if err == nil {
			io.Copy(&buf, rc)
			rc.Close()
		}
	}

	return buf.String()
}

// RawHTTPResponse format the http.Response to string.
// Notice, the order of headers is not strictly consistent
func RawHTTPResponse(resp *http.Response) string {
	var buf strings.Builder

	buf.WriteString(resp.Proto + " " + resp.Status + "\r\n")

	resp.Header.Write(&buf)
	buf.WriteString("\r\n")

	io.Copy(&buf, resp.Body)

	return buf.String()
}
