package zhttp

import (
	"net/http/cookiejar"
)

// Session is a client used to send http requests.
// Unlike Zhttp, it handle session for all requests
type Session struct {
	*Zhttp
	CookieJar *cookiejar.Jar
}

func (s *Session) Get(url string, options *ReqOptions) (*Response, error) {
	return s.doRequest("GET", url, options, s.CookieJar)
}

func (s *Session) Post(url string, options *ReqOptions) (*Response, error) {
	return s.doRequest("POST", url, options, s.CookieJar)
}

func (s *Session) Head(url string, options *ReqOptions) (*Response, error) {
	return s.doRequest("HEAD", url, options, s.CookieJar)
}

func (s *Session) Put(url string, options *ReqOptions) (*Response, error) {
	return s.doRequest("PUT", url, options, s.CookieJar)
}

func (s *Session) Delete(url string, options *ReqOptions) (*Response, error) {
	return s.doRequest("DELETE", url, options, s.CookieJar)
}

func (s *Session) Patch(url string, options *ReqOptions) (*Response, error) {
	return s.doRequest("PATCH", url, options, s.CookieJar)
}

func (s *Session) Options(url string, options *ReqOptions) (*Response, error) {
	return s.doRequest("OPTIONS", url, options, s.CookieJar)
}

func (s *Session) Request(method string, url string, options *ReqOptions) (*Response, error) {
	return s.doRequest(method, url, options, s.CookieJar)
}
