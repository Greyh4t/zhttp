package zhttp

import (
	"net/http"
	"net/http/cookiejar"

	"golang.org/x/net/publicsuffix"
)

type Session struct {
	CookieJar http.CookieJar
	Cookies   map[string]string
}

func NewSession(cookies map[string]string) *Session {
	s := new(Session)
	s.CookieJar, _ = cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	s.Cookies = cookies
	return s
}

func (s *Session) Get(url string, ro *RequestOptions) (*Response, error) {
	s.mergeCookies(ro)
	return doRequest("GET", url, ro, s.CookieJar)
}

func (s *Session) Post(url string, ro *RequestOptions) (*Response, error) {
	s.mergeCookies(ro)
	return doRequest("POST", url, ro, s.CookieJar)
}

func (s *Session) Head(url string, ro *RequestOptions) (*Response, error) {
	s.mergeCookies(ro)
	return doRequest("HEAD", url, ro, s.CookieJar)
}

func (s *Session) Put(url string, ro *RequestOptions) (*Response, error) {
	s.mergeCookies(ro)
	return doRequest("PUT", url, ro, s.CookieJar)
}

func (s *Session) Delete(url string, ro *RequestOptions) (*Response, error) {
	s.mergeCookies(ro)
	return doRequest("DELETE", url, ro, s.CookieJar)
}

func (s *Session) Patch(url string, ro *RequestOptions) (*Response, error) {
	s.mergeCookies(ro)
	return doRequest("PATCH", url, ro, s.CookieJar)
}

func (s *Session) Options(url string, ro *RequestOptions) (*Response, error) {
	s.mergeCookies(ro)
	return doRequest("OPTIONS", url, ro, s.CookieJar)
}

func (s *Session) Request(method string, url string, ro *RequestOptions) (*Response, error) {
	s.mergeCookies(ro)
	return doRequest(method, url, ro, s.CookieJar)
}

func (s *Session) mergeCookies(ro *RequestOptions) {
	if len(s.Cookies) > 0 {
		if ro.Cookies == nil {
			ro.Cookies = map[string]string{}
		}
		for k, v := range s.Cookies {
			if _, ok := ro.Cookies[k]; !ok {
				ro.Cookies[k] = v
			}
		}
	}
}
