package zhttp

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"
)

// ZBody is a wrapper for http.Response.ZBody
type ZBody struct {
	rawBody   io.ReadCloser
	buf       bytes.Buffer
	bufCached bool
	Err       error
}

// Read is the implementation of the reader interface
func (b *ZBody) Read(p []byte) (int, error) {
	if b.Err != nil {
		return 0, b.Err
	}

	return b.rawBody.Read(p)
}

// ReadN read and return n byte of body, and cache them
func (b *ZBody) ReadN(n int64) []byte {
	if b.Err != nil {
		return nil
	}

	lr := io.LimitReader(b.rawBody, n)
	tr := io.TeeReader(lr, &(b.buf))

	data, err := ioutil.ReadAll(tr)
	if err != nil && err != io.EOF {
		b.Err = err
		b.ClearCache()
		b.rawBody.Close()
		return nil
	}

	return data
}

// fillBuffer cache the body content – this is largely used for .String() and .Bytes()
func (b *ZBody) fillBuffer() {
	if b.bufCached {
		return
	}

	_, err := io.Copy(&b.buf, b.rawBody)
	b.bufCached = true

	if err != nil && err != io.EOF {
		b.Err = err
		b.ClearCache()
	}

	b.rawBody.Close()
}

// String return the body in string type
func (b *ZBody) String() string {
	if b.Err != nil {
		return ""
	}

	b.fillBuffer()

	return b.buf.String()
}

// Bytes return the body with []byte type
func (b *ZBody) Bytes() []byte {
	if b.Err != nil {
		return nil
	}

	b.fillBuffer()

	if b.buf.Len() == 0 {
		return nil
	}

	return b.buf.Bytes()
}

// Close close the body. Must be called when the response is used
func (b *ZBody) Close() error {
	if b.Err != nil {
		return b.Err
	}

	return b.rawBody.Close()
}

// ClearCache clear the cache of body
func (b *ZBody) ClearCache() {
	if b.buf.Len() > 0 {
		b.buf.Reset()
	}
}

// Headers is a wrapper for http.Header
type Headers map[string][]string

// String return a header in wire format.
func (h Headers) String() string {
	if h == nil {
		return ""
	}

	var buf strings.Builder
	http.Header(h).Write(&buf)
	return buf.String()
}

// Get gets the value associated with the given key. If
// there are no values associated with the key, Get returns "".
// multiple header fields with the same name will be join with ", ".
// It is case insensitive; textproto.CanonicalMIMEHeaderKey is
// used to canonicalize the provided key. To access multiple
// values of a key, or to use non-canonical keys, access the
// map directly.
func (h Headers) Get(key string) string {
	v := http.Header(h).Values(key)
	if len(v) == 0 {
		return ""
	}

	return strings.Join(v, ", ")
}

// Has will return information about whether a response header
// with the given name exists. If not exist, Has returns false.
// It is case insensitive;
func (h Headers) Has(key string) bool {
	if h == nil {
		return false
	}

	_, ok := h[http.CanonicalHeaderKey(key)]
	return ok
}

// Cookies is a wrapper for []*http.Cookie
type Cookies []*http.Cookie

// String return the cookies in string type.
// like key1=value1; key2=value2
func (c Cookies) String() string {
	if len(c) == 0 {
		return ""
	}

	var buf strings.Builder
	for i, cookie := range c {
		buf.WriteString(cookie.Name)
		buf.WriteRune('=')
		buf.WriteString(cookie.Value)
		if i < len(c)-1 {
			buf.WriteString("; ")
		}
	}

	return buf.String()
}

// Get gets the cookie value with the given name. If
// there are no values associated with the name, Get returns "".
func (c Cookies) Get(name string) string {
	for _, cookie := range c {
		if cookie.Name == name {
			return cookie.Value
		}
	}

	return ""
}

// Has will return whether the specified cookie is set in response.
func (c Cookies) Has(name string) bool {
	for _, cookie := range c {
		if cookie.Name == name {
			return true
		}
	}

	return false
}

// Response is a wrapper for *http.Response
type Response struct {
	StatusCode    int
	Status        string
	ContentLength int64
	Headers       Headers
	Body          *ZBody
	RawResponse   *http.Response
	cookies       Cookies
}

// Cookies parses and returns the cookies set in the Set-Cookie headers.
func (resp *Response) Cookies() Cookies {
	if resp.cookies == nil {
		resp.cookies = resp.RawResponse.Cookies()
	}

	return resp.cookies
}

// Ok validates that the server returned a 2xx code.
func (resp *Response) Ok() bool {
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// Err returns the first non-EOF error that was encountered by read body.
func (resp *Response) Err() error {
	return resp.Body.Err
}

// Close close the http response body.
func (resp *Response) Close() error {
	resp.Body.ClearCache()
	return resp.Body.Close()
}

// DumpRequest format the last http.Request to string.
// Notice, the order of headers is not strictly consistent
func (resp *Response) DumpRequest() string {
	var buf strings.Builder
	req := resp.RawResponse.Request.Clone(context.Background())

	data, _ := httputil.DumpRequestOut(req, false)
	buf.Write(data)

	if req.GetBody != nil {
		rc, err := req.GetBody()
		if err == nil {
			io.Copy(&buf, rc)
			rc.Close()
		}
	}

	return buf.String()
}

// DumpResponse format the last http.Response to string.
// Notice, the order of headers is not strictly consistent
func (resp *Response) DumpResponse() string {
	var buf strings.Builder

	data, _ := httputil.DumpResponse(resp.RawResponse, false)
	buf.Write(data)

	buf.Write(resp.Body.Bytes())

	return buf.String()
}
