package tools

import (
	"net/http"
	"strings"
)

// GetCookie check all responses in the redirect and return the first matching url and cookie
func GetCookie(resp *http.Response, name string) (string, string, bool) {
	if resp == nil {
		return "", "", false
	}

	req := resp.Request
	cookies := resp.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == name {
			return req.URL.String(), cookie.Value, true
		}
	}

	return GetCookie(req.Response, name)
}

// GetHeader check all responses in the redirect and return the first matching url and header
func GetHeader(resp *http.Response, key string) (string, string, bool) {
	if resp == nil {
		return "", "", false
	}

	req := resp.Request
	values, ok := resp.Header[http.CanonicalHeaderKey(key)]
	if ok {
		return req.URL.String(), strings.Join(values, ", "), true
	}

	return GetHeader(req.Response, key)
}

// CookieFromRaw parses a cookie in string format to []*http.Cookie
func CookieFromRaw(rawCookie string, domain string) []*http.Cookie {
	list := strings.Split(rawCookie, ";")
	if len(list) == 0 {
		return nil
	}

	cookies := make([]*http.Cookie, len(list))
	for i, item := range list {
		pairs := strings.SplitN(strings.TrimSpace(item), "=", 2)
		cookie := &http.Cookie{
			Name:   pairs[0],
			Domain: domain,
		}

		if len(pairs) == 2 {
			cookie.Value = pairs[1]
		}

		cookies[i] = cookie
	}

	return cookies
}
