package zhttp

import (
	"github.com/Greyh4t/dnscache"
)

var dnsCache *dnscache.Resolver

func SetDnsCache(resolver *dnscache.Resolver) {
	dnsCache = resolver
}

func Request(method, url string, ro *RequestOptions) (*Response, error) {
	return doRequest(method, url, ro, nil)
}

func Get(url string, ro *RequestOptions) (*Response, error) {
	return doRequest("GET", url, ro, nil)
}

func Delete(url string, ro *RequestOptions) (*Response, error) {
	return doRequest("DELETE", url, ro, nil)
}

func Head(url string, ro *RequestOptions) (*Response, error) {
	return doRequest("HEAD", url, ro, nil)
}

func Patch(url string, ro *RequestOptions) (*Response, error) {
	return doRequest("PATCH", url, ro, nil)
}

func Post(url string, ro *RequestOptions) (*Response, error) {
	return doRequest("POST", url, ro, nil)
}

func Put(url string, ro *RequestOptions) (*Response, error) {
	return doRequest("PUT", url, ro, nil)
}

func Options(url string, ro *RequestOptions) (*Response, error) {
	return doRequest("OPTIONS", url, ro, nil)
}
