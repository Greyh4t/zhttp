package zhttp

import (
	"net/url"
	"strings"
)

var quoteEscaper = strings.NewReplacer(`\`, `\\`, `"`, `\"`)

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

// MustProxy convert scheme and url string to map[string]*url.URL.
// If there have any error, will panic
func MustProxy(proxies map[string]string) map[string]*url.URL {
	if len(proxies) > 0 {
		proxiesMap := map[string]*url.URL{}
		for scheme, proxyURL := range proxies {
			urlObj, err := url.Parse(proxyURL)
			if err != nil {
				panic(err)
			}
			proxiesMap[scheme] = urlObj
		}
		return proxiesMap
	}
	return nil
}

// CookieMapFromRaw parse a cookie in string format to map[string]string
func CookieMapFromRaw(rawCookie string) map[string]string {
	list := strings.Split(rawCookie, ";")
	if len(list) == 0 {
		return nil
	}

	cookies := make(map[string]string, len(list))
	for _, item := range list {
		pairs := strings.SplitN(strings.TrimSpace(item), "=", 2)

		if len(pairs) == 2 {
			cookies[pairs[0]] = pairs[1]
		} else {
			cookies[pairs[0]] = ""
		}
	}

	return cookies
}
