package zhttp

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/textproto"
	"strings"
)

// Headers is a wrapper for http.Header
type headers struct {
	header http.Header
}

// String return a header in wire format.
func (h *headers) String() string {
	var buf strings.Builder
	h.header.Write(&buf)
	return buf.String()
}

// Get gets the value associated with the given key. If
// there are no values associated with the key, Get returns "".
// multiple header fields with the same name will be join with ", ".
// It is case insensitive; textproto.CanonicalMIMEHeaderKey is
// used to canonicalize the provided key. To access multiple
// values of a key, or to use non-canonical keys, access the
// map directly.
func (h *headers) Get(key string) string {
	v, ok := h.header[textproto.CanonicalMIMEHeaderKey(key)]
	if !ok {
		return ""
	}
	return strings.Join(v, ", ")
}

// Has will return information about whether a response header
// with the given name exists. If not exist, Has returns false.
// It is case insensitive;
func (h *headers) Has(key string) bool {
	_, ok := h.header[textproto.CanonicalMIMEHeaderKey(key)]
	return ok
}

// Cookies is a wrapper for []*http.Cookie
type cookies struct {
	cookies []*http.Cookie
	get     func() []*http.Cookie
}

func (c *cookies) parse() {
	if c.cookies == nil {
		c.cookies = c.get()
	}
}

// String return the cookies in string type.
// like key1=value1; key2=value2
func (c *cookies) String() string {
	c.parse()

	var buf strings.Builder
	for i, cookie := range c.cookies {
		buf.WriteString(cookie.Name)
		buf.WriteRune('=')
		buf.WriteString(cookie.Value)
		if i < len(c.cookies)-1 {
			buf.WriteString("; ")
		}
	}
	return buf.String()
}

// Get gets the cookie value with the given name. If
// there are no values associated with the name, Get returns "".
func (c *cookies) Get(name string) string {
	c.parse()

	for _, cookie := range c.cookies {
		if cookie.Name == name {
			return cookie.Value
		}
	}
	return ""
}

// Has will return whether the specified cookie is set in response.
func (c *cookies) Has(name string) bool {
	c.parse()

	for _, cookie := range c.cookies {
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
	Headers       headers
	Cookies       cookies
	RawResponse   *http.Response
	Error         error
}

// String return the body in string type
func (resp *Response) String() string {
	data, err := ioutil.ReadAll(resp.RawResponse.Body)
	resp.Error = err
	return string(data)
}

// Byte return the body with []byte type
func (resp *Response) Byte() []byte {
	data, err := ioutil.ReadAll(resp.RawResponse.Body)
	resp.Error = err
	return data
}

// ReadN read and return n byte of body
func (resp *Response) ReadN(n int64) []byte {
	body, err := ioutil.ReadAll(io.LimitReader(resp.RawResponse.Body, n))
	resp.Error = err
	return body
}

// Close close the body. Must be called when the response is used
func (resp *Response) Close() error {
	return resp.RawResponse.Body.Close()
}
