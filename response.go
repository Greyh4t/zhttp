package zhttp

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type Response struct {
	StatusCode    int
	Status        string
	ContentLength int64
	RawResponse   *http.Response
	Err           error
}

// String return the body in string type
func (resp *Response) String() string {
	data, err := ioutil.ReadAll(resp.RawResponse.Body)
	if err != nil {
		resp.Err = err
		return ""
	}
	return string(data)
}

// Byte return the body with []byte type
func (resp *Response) Byte() []byte {
	data, err := ioutil.ReadAll(resp.RawResponse.Body)
	if err != nil {
		resp.Err = err
		return nil
	}
	return data
}

// ReadN read and return n byte of body
func (resp *Response) ReadN(n int64) []byte {
	body, _ := ioutil.ReadAll(io.LimitReader(resp.RawResponse.Body, n))
	return body
}

// Close close the body
func (resp *Response) Close() error {
	return resp.RawResponse.Body.Close()
}

/* RawHeaders return the headers in string type
like this form
header1: value1,value11
header2: value2
*/
func (resp *Response) RawHeaders() string {
	var rawHeader string
	for k, v := range resp.RawResponse.Header {
		rawHeader += k + ": " + strings.Join(v, ",") + "\r\n"
	}
	return strings.TrimSuffix(rawHeader, "\r\n")
}

// HeadersMap return the headers in a map
func (resp *Response) HeadersMap() map[string]string {
	headers := map[string]string{}
	for k, v := range resp.RawResponse.Header {
		headers[k] = strings.Join(v, ",")
	}
	return headers
}

// GetHeader return a specific header
// If header not exist, return empty string and false
func (resp *Response) GetHeader(name string) (string, bool) {
	for k, v := range resp.RawResponse.Header {
		if k == name {
			return strings.Join(v, ","), true
		}
	}
	return "", false
}

// RawCookies return the headers in string type
// like key1=value1; key2=value2
func (resp *Response) RawCookies() string {
	var rawCookie string
	for _, cookie := range resp.RawResponse.Cookies() {
		rawCookie += cookie.Name + "=" + cookie.Value + ";"
	}
	return strings.TrimSuffix(rawCookie, ";")
}

// CookiesMap return the cookies in a map
func (resp *Response) CookiesMap() map[string]string {
	cookies := map[string]string{}
	for _, cookie := range resp.RawResponse.Cookies() {
		cookies[cookie.Name] = cookie.Value
	}
	return cookies
}

// GetCookie return a specific cookie
// If cookie not exist, return empty string and false
func (resp *Response) GetCookie(name string) (string, bool) {
	for _, cookie := range resp.RawResponse.Cookies() {
		if cookie.Name == name {
			return cookie.Value, true
		}
	}
	return "", false
}

// RawRequest return the last request in string
// Notice, the order of the headers is not correct
func (resp *Response) RawRequest() string {
	var rawRequest bytes.Buffer
	rawRequest.WriteString(resp.RawResponse.Request.Method + " " + resp.RawResponse.Request.URL.RequestURI() +
		" " + resp.RawResponse.Request.Proto + "\r\n")

	if resp.RawResponse.Request.Host != "" {
		rawRequest.WriteString("Host: " + resp.RawResponse.Request.Host + "\r\n")
	} else {
		rawRequest.WriteString("Host: " + resp.RawResponse.Request.URL.Host + "\r\n")
	}

	for key, val := range resp.RawResponse.Request.Header {
		rawRequest.WriteString(key + ": " + val[0] + "\r\n")
	}

	rawRequest.WriteString("\r\n")
	rawRequest.Write(resp.reqBody())

	return rawRequest.String()
}

func (resp *Response) reqBody() []byte {
	if resp.RawResponse.Request.GetBody != nil {
		rc, err := resp.RawResponse.Request.GetBody()
		if err == nil {
			body, err := ioutil.ReadAll(rc)
			rc.Close()
			if err == nil {
				return body
			}
		}
	}
	return nil
}
