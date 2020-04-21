package zhttp

import (
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

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
	RawResponse   *http.Response
	Error         error
	cookies       Cookies
}

// Cookies parses and returns the cookies set in the Set-Cookie headers.
func (resp *Response) Cookies() Cookies {
	if resp.cookies == nil {
		resp.cookies = resp.RawResponse.Cookies()
	}

	return resp.cookies
}

// String return the body in string type
func (resp *Response) String() string {
	data := resp.read(resp.RawResponse.Body)
	if data == nil {
		return ""
	}

	return string(data)
}

// Byte return the body with []byte type
func (resp *Response) Byte() []byte {
	data := resp.read(resp.RawResponse.Body)
	return data
}

// ReadN read and return n byte of body
func (resp *Response) ReadN(n int64) []byte {
	data := resp.read(io.LimitReader(resp.RawResponse.Body, n))
	return data
}

// Close close the body. Must be called when the response is used
func (resp *Response) Close() error {
	if resp.Error != nil {
		return resp.Error
	}

	return resp.RawResponse.Body.Close()
}

func (resp *Response) read(body io.Reader) []byte {
	if resp.Error != nil {
		return nil
	}

	data, err := ioutil.ReadAll(body)
	if err != nil {
		resp.Error = err
		resp.RawResponse.Body.Close()
		return nil
	}

	return data
}
