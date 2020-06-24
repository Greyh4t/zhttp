package zhttp

import (
	"net/http"
	"net/http/cookiejar"
	"runtime"
	"time"

	"github.com/greyh4t/dnscache"
	"golang.org/x/net/publicsuffix"
)

type Zhttp struct {
	options   *HTTPOptions
	dnsCache  *dnscache.Cache
	transport *http.Transport
}

// New generate an *Zhttp client to send request
func New(options *HTTPOptions) *Zhttp {
	z := &Zhttp{options: options}
	if z.options == nil {
		z.options = &HTTPOptions{}
	}

	var cache *dnscache.Cache
	if z.options.DNSCacheExpire > 0 {
		if z.options.DNSServer != "" {
			cache = dnscache.NewWithServer(z.options.DNSCacheExpire, z.options.DNSServer)
		} else {
			cache = dnscache.New(z.options.DNSCacheExpire)
		}
		z.dnsCache = cache
	}

	z.transport = createTransport(z.options, cache)

	ensureResourcesFinalized(z, z.dnsCache != nil)

	return z
}

// NewWithDNSCache generate an *Zhttp client that uses an external DNSCache.
// This will ignore HTTPOptions.DNSCacheExpire and HTTPOptions.DNSServer
func NewWithDNSCache(options *HTTPOptions, cache *dnscache.Cache) *Zhttp {
	z := &Zhttp{options: options}
	if z.options == nil {
		z.options = &HTTPOptions{}
	}

	if cache != nil {
		z.dnsCache = cache
	}

	z.transport = createTransport(z.options, cache)

	ensureResourcesFinalized(z, false)

	return z
}

func ensureResourcesFinalized(zhttp *Zhttp, finalizeDNSCache bool) {
	runtime.SetFinalizer(zhttp, func(z *Zhttp) {
		z.transport.CloseIdleConnections()
		if finalizeDNSCache {
			z.dnsCache.Close()
		}
	})
}

// NewSession generate an client that will handle session for all requests
func (z *Zhttp) NewSession() *Session {
	s := &Session{z: z}
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

var defaultZ *Zhttp

// InitDefaultClient initialization the default zhttp client with options
func InitDefaultClient(options *HTTPOptions) {
	defaultZ = New(options)
}

func prepareDefaultZ() {
	if defaultZ == nil {
		defaultZ = New(&HTTPOptions{
			Timeout: time.Second * 30,
		})
	}
}

// NewSession generate an default client that will handle session for all requests
func NewSession() *Session {
	prepareDefaultZ()
	return defaultZ.NewSession()
}

func Request(method, url string, options *ReqOptions) (*Response, error) {
	prepareDefaultZ()
	return defaultZ.doRequest(method, url, options, nil)
}

func Get(url string, options *ReqOptions) (*Response, error) {
	prepareDefaultZ()
	return defaultZ.doRequest("GET", url, options, nil)
}

func Delete(url string, options *ReqOptions) (*Response, error) {
	prepareDefaultZ()
	return defaultZ.doRequest("DELETE", url, options, nil)
}

func Head(url string, options *ReqOptions) (*Response, error) {
	prepareDefaultZ()
	return defaultZ.doRequest("HEAD", url, options, nil)
}

func Patch(url string, options *ReqOptions) (*Response, error) {
	prepareDefaultZ()
	return defaultZ.doRequest("PATCH", url, options, nil)
}

func Post(url string, options *ReqOptions) (*Response, error) {
	prepareDefaultZ()
	return defaultZ.doRequest("POST", url, options, nil)
}

func Put(url string, options *ReqOptions) (*Response, error) {
	prepareDefaultZ()
	return defaultZ.doRequest("PUT", url, options, nil)
}

func Options(url string, options *ReqOptions) (*Response, error) {
	prepareDefaultZ()
	return defaultZ.doRequest("OPTIONS", url, options, nil)
}
