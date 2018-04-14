package zhttp

import (
	"github.com/Greyh4t/dnscache"
)

var dnsCache *dnscache.Resolver

func SetDnsCache(resolver *dnscache.Resolver) {
	dnsCache = resolver
}

func Request(method, url string, ro *RequestOptions) (*Response, error) {
	return doRequest(method, url, ro)
}

func Get(url string, ro *RequestOptions) (*Response, error) {
	return doRequest("GET", url, ro)
}

func Delete(url string, ro *RequestOptions) (*Response, error) {
	return doRequest("DELETE", url, ro)
}

func Head(url string, ro *RequestOptions) (*Response, error) {
	return doRequest("HEAD", url, ro)
}

func Patch(url string, ro *RequestOptions) (*Response, error) {
	return doRequest("PATCH", url, ro)
}

func Post(url string, ro *RequestOptions) (*Response, error) {
	return doRequest("POST", url, ro)
}

func Put(url string, ro *RequestOptions) (*Response, error) {
	return doRequest("PUT", url, ro)
}

func Options(url string, ro *RequestOptions) (*Response, error) {
	return doRequest("OPTIONS", url, ro)
}
