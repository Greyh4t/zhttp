package zhttp

import (
	"github.com/greyh4t/dnscache"
	"golang.org/x/net/publicsuffix"
	"net/http"
	"net/http/cookiejar"
	"runtime"
)

type Zhttp struct {
	options   *HttpOptions
	dnsCache  *dnscache.Resolver
	transport *http.Transport
}

// New generate an *Zhttp client to send request
func New(options *HttpOptions) *Zhttp {
	z := &Zhttp{options: options}
	if z.options == nil {
		z.options = &HttpOptions{}
	}

	if z.options.DNSCacheExpire > 0 {
		if z.options.DNSServer != "" {
			z.dnsCache = dnscache.NewCustomServer(z.options.DNSCacheExpire, z.options.DNSServer)
		} else {
			z.dnsCache = dnscache.New(z.options.DNSCacheExpire)
		}
		ensureDNSCacheFinalized(z.dnsCache)
	}

	z.transport = z.createTransport(z.options)
	ensureTransporterFinalized(z.transport)
	return z
}

// NewWithDNSCache generate an *Zhttp client that uses an external DNSCache
// This function will ignore HttpOptions.DNSCacheExpire and HttpOptions.DNSServer
func NewWithDNSCache(options *HttpOptions, cache *dnscache.Resolver) *Zhttp {
	z := &Zhttp{options: options}
	if z.options == nil {
		z.options = &HttpOptions{}
	}

	if cache != nil {
		z.dnsCache = cache
	}

	z.transport = z.createTransport(z.options)
	ensureTransporterFinalized(z.transport)
	return z
}

func ensureDNSCacheFinalized(resolver *dnscache.Resolver) {
	runtime.SetFinalizer(&resolver, func(resolver **dnscache.Resolver) {
		(*resolver).Close()
	})
}

func ensureTransporterFinalized(httpTransport *http.Transport) {
	runtime.SetFinalizer(&httpTransport, func(transportInt **http.Transport) {
		(*transportInt).CloseIdleConnections()
	})
}

// NewSession generate an client that will handle session for all request
func (z *Zhttp) NewSession() *Session {
	s := &Session{Zhttp: z}
	s.CookieJar, _ = cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	return s
}

func (z *Zhttp) Request(method, url string, options *ReqOptions) (*Response, error) {
	return z.doRequest(method, url, options, nil)
}

func (z *Zhttp) Get(url string, options *ReqOptions) (*Response, error) {
	return z.doRequest("GET", url, options, nil)
}

func (z *Zhttp) Delete(url string, options *ReqOptions) (*Response, error) {
	return z.doRequest("DELETE", url, options, nil)
}

func (z *Zhttp) Head(url string, options *ReqOptions) (*Response, error) {
	return z.doRequest("HEAD", url, options, nil)
}

func (z *Zhttp) Patch(url string, options *ReqOptions) (*Response, error) {
	return z.doRequest("PATCH", url, options, nil)
}

func (z *Zhttp) Post(url string, options *ReqOptions) (*Response, error) {
	return z.doRequest("POST", url, options, nil)
}

func (z *Zhttp) Put(url string, options *ReqOptions) (*Response, error) {
	return z.doRequest("PUT", url, options, nil)
}

func (z *Zhttp) Options(url string, options *ReqOptions) (*Response, error) {
	return z.doRequest("OPTIONS", url, options, nil)
}
