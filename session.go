package zhttp

import (
	"net/http/cookiejar"
)

// Session is a client used to send http requests.
// Unlike Zhttp, it handle session for all requests
type Session struct {
	z         *Zhttp
	CookieJar *cookiejar.Jar
}

func (s *Session) Get(url string, options *ReqOptions) (*Response, error) {
	return s.z.doRequest("GET", url, options, s.CookieJar)
}

func (s *Session) Post(url string, options *ReqOptions) (*Response, error) {
	return s.z.doRequest("POST", url, options, s.CookieJar)
}

func (s *Session) Head(url string, options *ReqOptions) (*Response, error) {
	return s.z.doRequest("HEAD", url, options, s.CookieJar)
}

func (s *Session) Put(url string, options *ReqOptions) (*Response, error) {
	return s.z.doRequest("PUT", url, options, s.CookieJar)
}

func (s *Session) Delete(url string, options *ReqOptions) (*Response, error) {
	return s.z.doRequest("DELETE", url, options, s.CookieJar)
}

func (s *Session) Patch(url string, options *ReqOptions) (*Response, error) {
	return s.z.doRequest("PATCH", url, options, s.CookieJar)
}

func (s *Session) Options(url string, options *ReqOptions) (*Response, error) {
	return s.z.doRequest("OPTIONS", url, options, s.CookieJar)
}

func (s *Session) Request(method string, url string, options *ReqOptions) (*Response, error) {
	return s.z.doRequest(method, url, options, s.CookieJar)
}
