package zhttp

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
)

var quoteEscaper = strings.NewReplacer(`\`, `\\`, `"`, `\"`)

type FileUpload struct {
	// Filename is the name of the file that you wish to upload. We use this to guess the mimetype as well as pass it onto the server
	FileName string

	// FileContents is happy as long as you pass it a io.ReadCloser (which most file use anyways)
	FileContents io.ReadCloser

	// FieldName is form field name
	FieldName string

	// FileMime represents which mimetime should be sent along with the file.
	// When empty, defaults to application/octet-stream
	FileMime string
}

type RequestOptions struct {
	// Data is a map of key values that will eventually convert into the
	// query string of a GET request or the body of a POST request.
	Data map[string]string

	// Params is a map of query strings that may be used within a GET request
	Params map[string]string

	// Files is where you can include files to upload. The use of this data
	// structure is limited to POST requests
	Files []FileUpload

	// JSON can be used when you wish to send JSON within the request body
	JSON interface{}

	// XML can be used if you wish to send XML within the request body
	XML interface{}

	// Headers if you want to add custom HTTP headers to the request,
	// this is your friend
	Headers map[string]string

	// InsecureSkipVerify is a flag that specifies if we should validate the
	// server's TLS certificate. It should be noted that Go's TLS verify mechanism
	// doesn't validate if a certificate has been revoked
	InsecureSkipVerify bool

	// DisableCompression will disable gzip compression on requests
	DisableCompression bool

	// Host allows you to set an arbitrary custom host
	Host string

	// Set custom dns resolution for the current request
	Hosts string

	ContentType string
	UserAgent   string

	// Auth allows you to specify a user name and password that you wish to
	// use when requesting the URL. It will use basic HTTP authentication
	// formatting the username and password in base64 the format is:
	// []string{username, password}
	Auth []string

	// Cookies is an array of `http.Cookie` that allows you to attach
	// cookies to your request
	Cookies map[string]string

	// Proxies is a map in the following format
	// *protocol* => proxy address e.g http => http://127.0.0.1:8080
	Proxies map[string]*url.URL

	// DialTimeout is the maximum amount of time a dial will wait for
	// a connect to complete.
	DialTimeout time.Duration

	// RequestTimeout is the maximum amount of time a whole request(include dial / request / redirect)
	// will wait.
	RequestTimeout time.Duration

	DisableRedirect bool

	// RequestBody allows you to put anything matching an `io.Reader` into the request
	// this option will take precedence over any other request option specified
	//RequestBody io.Reader

	RawCookie string
	RawQuery  string
	RawData   string
	IsAjax    bool
}

func (ro *RequestOptions) closeFiles() {
	for _, f := range ro.Files {
		f.FileContents.Close()
	}
}

func (ro RequestOptions) proxySettings(req *http.Request) (*url.URL, error) {
	if len(ro.Proxies) > 0 {
		// There was a proxy specified – do we support the protocol?
		if p, ok := ro.Proxies[req.URL.Scheme]; ok {
			return p, nil
		}
	}

	// Proxies were specified but not for any protocol that we use
	return http.ProxyFromEnvironment(req)
}

func FileUploadFromDisk(fieldName, filePath string) ([]FileUpload, error) {
	filePath = filepath.ToSlash(filepath.Clean(filePath))
	fd, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	_, fileName := filepath.Split(filePath)
	return []FileUpload{
		FileUpload{
			FieldName:    fieldName,
			FileName:     fileName,
			FileContents: fd,
		},
	}, nil
}

func createTransport(ro RequestOptions) *http.Transport {
	transport := &http.Transport{
		MaxIdleConns:          100,
		IdleConnTimeout:       5 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		Proxy:                 ro.proxySettings,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: ro.InsecureSkipVerify},
		DisableCompression:    ro.DisableCompression,
		DisableKeepAlives:     true,
	}

	if dnsCache != nil {
		transport.Dial = func(network, address string) (net.Conn, error) {
			host, port, _ := net.SplitHostPort(address)
			ip, err := dnsCache.FetchOneString(host)
			if err != nil {
				return nil, err
			}
			conn, err := net.DialTimeout(network, net.JoinHostPort(ip, port), ro.DialTimeout)
			if err != nil {
				return nil, err
			}
			return newTimeoutConn(conn, ro.RequestTimeout), nil
		}
	} else {
		transport.Dial = func(network, address string) (net.Conn, error) {
			conn, err := net.DialTimeout(network, address, ro.DialTimeout)
			if err != nil {
				return nil, err
			}
			return newTimeoutConn(conn, ro.RequestTimeout), nil
		}
	}

	return transport
}

func buildClient(ro RequestOptions, cookieJar http.CookieJar) *http.Client {
	if ro.DialTimeout <= 0 {
		if ro.RequestTimeout > 0 && ro.RequestTimeout < 10*time.Second {
			ro.DialTimeout = ro.RequestTimeout
		} else {
			ro.DialTimeout = 10 * time.Second
		}
	}

	// The function does not return an error ever... so we are just ignoring it
	if cookieJar == nil {
		cookieJar, _ = cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	}

	client := &http.Client{
		Jar:       cookieJar,
		Transport: createTransport(ro),
		Timeout:   ro.RequestTimeout,
	}

	if ro.DisableRedirect {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return client
}

func buildRequest(method, urlStr string, ro *RequestOptions) (*http.Request, error) {
	if ro.RawData != "" {
		return http.NewRequest(method, urlStr, strings.NewReader(ro.RawData))
	}

	if ro.JSON != nil {
		return createBasicJSONRequest(method, urlStr, ro)
	}

	if ro.XML != nil {
		return createBasicXMLRequest(method, urlStr, ro)
	}

	if ro.Files != nil {
		return createFileUploadRequest(method, urlStr, ro)
	}

	if ro.Data != nil {
		return createBasicRequest(method, urlStr, ro)
	}

	return http.NewRequest(method, urlStr, nil)
}

func createBasicJSONRequest(method, urlStr string, ro *RequestOptions) (*http.Request, error) {
	var reader io.Reader
	switch v := ro.JSON.(type) {
	case string:
		reader = strings.NewReader(v)
	case []byte:
		reader = bytes.NewReader(v)
	default:
		jsonB, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(jsonB)
	}

	req, err := http.NewRequest(method, urlStr, reader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func createBasicXMLRequest(method, urlStr string, ro *RequestOptions) (*http.Request, error) {
	var reader io.Reader
	switch v := ro.XML.(type) {
	case string:
		reader = strings.NewReader(v)
	case []byte:
		reader = bytes.NewReader(v)
	default:
		xmlB, err := xml.Marshal(v)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(xmlB)
	}

	req, err := http.NewRequest(method, urlStr, reader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/xml")

	return req, nil
}

func createFileUploadRequest(method, urlStr string, ro *RequestOptions) (*http.Request, error) {
	if method == "POST" {
		return createMultiPartPostRequest(method, urlStr, ro)
	}

	// This may be a PUT or PATCH request so we will just put the raw
	// io.ReadCloser in the request body
	// and guess the MIME type from the file name

	// At the moment, we will only support 1 file upload as a time
	// when uploading using PUT or PATCH

	req, err := http.NewRequest(method, urlStr, ro.Files[0].FileContents)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", mime.TypeByExtension(filepath.Ext(ro.Files[0].FileName)))

	return req, nil
}

func createMultiPartPostRequest(method, urlStr string, ro *RequestOptions) (*http.Request, error) {
	body := &bytes.Buffer{}

	multipartWriter := multipart.NewWriter(body)

	for i, f := range ro.Files {
		if f.FileContents == nil {
			return nil, errors.New("Pointer FileContents cannot be nil")
		}

		fieldName := f.FieldName

		if fieldName == "" {
			if len(ro.Files) > 1 {
				fieldName = "file" + strconv.Itoa(i+1)
			} else {
				fieldName = "file"
			}
		}

		var writer io.Writer
		var err error

		if f.FileMime != "" {
			h := make(textproto.MIMEHeader)
			h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, escapeQuotes(fieldName), escapeQuotes(f.FileName)))
			h.Set("Content-Type", f.FileMime)
			writer, err = multipartWriter.CreatePart(h)
		} else {
			writer, err = multipartWriter.CreateFormFile(fieldName, f.FileName)
		}

		if err != nil {
			return nil, err
		}

		if _, err = io.Copy(writer, f.FileContents); err != nil && err != io.EOF {
			return nil, err
		}
	}

	// Populate the other parts of the form (if there are any)
	for key, value := range ro.Data {
		multipartWriter.WriteField(key, value)
	}

	if err := multipartWriter.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, urlStr, body)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", multipartWriter.FormDataContentType())

	return req, err
}

func createBasicRequest(method, urlStr string, ro *RequestOptions) (*http.Request, error) {
	req, err := http.NewRequest(method, urlStr, strings.NewReader(encodePostValues(ro.Data)))

	if err != nil {
		return nil, err
	}

	// The content type must be set to a regular form
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req, nil
}

func doRequest(method, urlStr string, ro *RequestOptions, cookieJar http.CookieJar) (*Response, error) {
	if ro == nil {
		ro = &RequestOptions{}
	}

	defer ro.closeFiles()

	urlStr, err := buildURL(urlStr, ro)
	if err != nil {
		return nil, err
	}

	req, err := buildRequest(method, urlStr, ro)
	if err != nil {
		return nil, err
	}

	parseHosts(req, ro)
	addCookies(req, ro)
	addHeaders(req, ro)

	client := buildClient(*ro, cookieJar)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return &Response{resp}, nil
}

// buildURLParams returns a URL with all of the params
// Note: This function will override current URL params if they contradict what is provided in the map
// That is what the "magic" is on the last line
func buildURL(urlStr string, ro *RequestOptions) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	if ro.RawQuery != "" {
		parsedURL.RawQuery = ro.RawQuery
	} else {
		if parsedURL.RawQuery != "" {
			parsedURL.RawQuery = escapeRawQuery(parsedURL.RawQuery)
		}
		if len(ro.Params) > 0 {
			query := url.Values{}
			for key, value := range ro.Params {
				query.Set(key, value)
			}
			if parsedURL.RawQuery != "" {
				parsedURL.RawQuery += "&" + query.Encode()
			} else {
				parsedURL.RawQuery = query.Encode()
			}
		}
	}

	return parsedURL.String(), nil
}

func parseHosts(req *http.Request, ro *RequestOptions) {
	if ro.Hosts != "" {
		req.Host = req.URL.Host
		port := req.URL.Port()
		if port != "" {
			req.URL.Host = ro.Hosts + ":" + port
		} else {
			req.URL.Host = ro.Hosts
		}
	}
}

// addHTTPHeaders adds any additional HTTP headers that need to be added are added here including:
// 1. Authorization Headers
// 2. Any other header requested
func addHeaders(req *http.Request, ro *RequestOptions) {
	req.Header.Set("User-Agent", "Zhttp/1.0")

	for key, value := range ro.Headers {
		req.Header.Set(key, value)
	}

	if ro.Host != "" {
		req.Host = ro.Host
	}

	if ro.Auth != nil {
		req.SetBasicAuth(ro.Auth[0], ro.Auth[1])
	}

	if ro.IsAjax {
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
	}

	if ro.ContentType != "" {
		req.Header.Set("Content-Type", ro.ContentType)
	}

	if ro.UserAgent != "" {
		req.Header.Set("User-Agent", ro.UserAgent)
	}
}

func addCookies(req *http.Request, ro *RequestOptions) {
	if ro.RawCookie != "" {
		req.Header.Set("Cookie", ro.RawCookie)
	}
	for k, v := range ro.Cookies {
		req.AddCookie(&http.Cookie{Name: k, Value: v})
	}
}

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

func encodePostValues(postValues map[string]string) string {
	urlValues := &url.Values{}

	for key, value := range postValues {
		urlValues.Set(key, value)
	}

	return urlValues.Encode() // This will sort all of the string values
}

func escapeRawQuery(s string) string {
	var r string
	for i := 0; i < len(s); {
		switch s[i] {
		case '%':
			if i+2 >= len(s) || !ishex(s[i+1]) || !ishex(s[i+2]) {
				r += string(int(s[i]))
				i++
			} else {
				c := unhex(s[i+1])<<4 | unhex(s[i+2])
				if shouldEscape(c) {
					r += s[i : i+3]
				} else {
					r += string(int(c))
				}
				i += 3
			}
		default:
			if shouldEscape(s[i]) {
				r += "%" + string(int("0123456789ABCDEF"[s[i]>>4])) + string(int("0123456789ABCDEF"[s[i]&15]))
			} else {
				r += string(int(s[i]))
			}
			i++
		}
	}

	return r
}

func ishex(c byte) bool {
	switch {
	case '0' <= c && c <= '9':
		return true
	case 'a' <= c && c <= 'f':
		return true
	case 'A' <= c && c <= 'F':
		return true
	}
	return false
}

func unhex(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}

func shouldEscape(c byte) bool {
	if 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || '0' <= c && c <= '9' {
		return false
	}
	switch c {
	case '!', '#', '$', '&', '\'', '(', ')', '*', '+', ',', '/', ':', ';', '=', '?', '@', '[', ']', '-', '.', '_', '~':
		return false
	}
	return true
}
