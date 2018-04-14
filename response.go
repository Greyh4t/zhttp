package zhttp

import (
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type Response struct {
	RawResponse *http.Response
}

func (self *Response) StatusCode() int {
	return self.RawResponse.StatusCode
}

func (self *Response) RawHeaders() string {
	var rawHeader string
	for k, v := range self.RawResponse.Header {
		rawHeader += k + ": " + strings.Join(v, ",") + "\r\n"
	}
	return strings.TrimSuffix(rawHeader, "\r\n")
}

func (self *Response) HeadersMap() map[string]string {
	headers := map[string]string{}
	for k, v := range self.RawResponse.Header {
		headers[k] = strings.Join(v, ",")
	}
	return headers
}

func (self *Response) HasHeader(headerName string) bool {
	for k, _ := range self.RawResponse.Header {
		if k == headerName {
			return true
		}
	}
	return false
}

func (self *Response) HasHeaderAndValue(headerName, headerVaule string) bool {
	for name, values := range self.RawResponse.Header {
		if name == headerName {
			for _, value := range values {
				if value == headerVaule {
					return true
				}
			}
		}
	}
	return false
}

func (self *Response) GetHeader(headerName string) string {
	for k, v := range self.RawResponse.Header {
		if k == headerName {
			return strings.Join(v, ",")
		}
	}
	return ""
}

func (self *Response) RawCookies() string {
	var rawCookie string
	for _, cookie := range self.RawResponse.Cookies() {
		rawCookie += cookie.Name + "=" + cookie.Value + ";"
	}
	return strings.TrimSuffix(rawCookie, ";")
}

func (self *Response) CookiesMap() map[string]string {
	cookies := map[string]string{}
	for _, cookie := range self.RawResponse.Cookies() {
		cookies[cookie.Name] = cookie.Value
	}
	return cookies
}

func (self *Response) HasCookieAndValue(cookieName, cookieVaule string) bool {
	for _, cookie := range self.RawResponse.Cookies() {
		if cookie.Name == cookieName && cookie.Value == cookieVaule {
			return true
		}
	}
	return false
}

func (self *Response) HasCookie(cookieName string) bool {
	for _, cookie := range self.RawResponse.Cookies() {
		if cookie.Name == cookieName {
			return true
		}
	}
	return false
}

func (self *Response) String() string {
	body, _ := ioutil.ReadAll(self.RawResponse.Body)
	return string(body)
}

func (self *Response) Byte() []byte {
	body, _ := ioutil.ReadAll(self.RawResponse.Body)
	return body
}

func (self *Response) ReadN(n int64) []byte {
	body, _ := ioutil.ReadAll(io.LimitReader(self.RawResponse.Body, n))
	return body
}

func (self *Response) RawRequest() string {
	rawRequest := self.RawResponse.Request.Method + " " + self.RawResponse.Request.URL.RequestURI() + " " + self.RawResponse.Request.Proto + "\r\n"
	host := self.RawResponse.Request.Host
	if host == "" {
		host = self.RawResponse.Request.URL.Host
	}
	rawRequest += "Host: " + host + "\r\n"
	for key, val := range self.RawResponse.Request.Header {
		rawRequest += key + ": " + val[0] + "\r\n"
	}
	rawRequest += "\r\n" + self.reqBody()

	return rawRequest
}

func (self *Response) reqBody() string {
	var body string
	if self.RawResponse.Request.GetBody != nil {
		b, err := self.RawResponse.Request.GetBody()
		if err == nil {
			buf, err := ioutil.ReadAll(b)
			b.Close()
			if err == nil {
				body = string(buf)
			}
		}
	}
	return body
}

func (self *Response) Close() error {
	return self.RawResponse.Body.Close()
}
