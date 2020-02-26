package zhttp

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"golang.org/x/net/publicsuffix"
)

// buildClient make a new client
func (z *Zhttp) buildClient(options *HTTPOptions, cookieJar http.CookieJar) *http.Client {
	if cookieJar == nil {
		cookieJar, _ = cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	}

	client := &http.Client{
		Transport: z.transport,
		Jar:       cookieJar,
		Timeout:   options.Timeout,
	}

	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		reqOptions, ok := req.Context().Value("options").(*ReqOptions)
		if (ok && reqOptions.DisableRedirect) || options.DisableRedirect {
			return http.ErrUseLastResponse
		}
		return nil
	}

	return client
}

// createTransport create a global *http.Transport for all http client
func (z *Zhttp) createTransport(options *HTTPOptions) *http.Transport {
	transport := http.DefaultTransport.(*http.Transport)
	transport.MaxIdleConnsPerHost = options.MaxIdleConnsPerHost
	transport.MaxConnsPerHost = options.MaxConnsPerHost
	transport.DisableKeepAlives = options.DisableKeepAlives
	transport.DisableCompression = options.DisableCompression

	if options.InsecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: options.InsecureSkipVerify}
	}
	if options.IdleConnTimeout > 0 {
		transport.IdleConnTimeout = options.IdleConnTimeout
	}
	if options.MaxIdleConns > 0 {
		transport.MaxIdleConns = options.MaxIdleConns
	}
	if options.TLSHandshakeTimeout > 0 {
		transport.TLSHandshakeTimeout = options.TLSHandshakeTimeout
	}

	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	if options.DialTimeout > 0 {
		dialer.Timeout = options.DialTimeout
	}
	if options.KeepAlive > 0 {
		dialer.KeepAlive = options.KeepAlive
	}

	transport.Proxy = func(req *http.Request) (*url.URL, error) {
		reqOptions, ok := req.Context().Value("options").(*ReqOptions)
		if ok && len(reqOptions.Proxies) > 0 {
			if p, ok := reqOptions.Proxies[req.URL.Scheme]; ok {
				return p, nil
			}
		} else if len(options.Proxies) > 0 {
			if p, ok := options.Proxies[req.URL.Scheme]; ok {
				return p, nil
			}
		}
		// get proxy from environment
		return http.ProxyFromEnvironment(req)
	}

	if z.dnsCache != nil {
		transport.DialContext = func(ctx context.Context, network string, address string) (net.Conn, error) {
			host, port, _ := net.SplitHostPort(address)
			ip, err := z.dnsCache.FetchOneString(host)
			if err != nil {
				return nil, err
			}
			return dialer.DialContext(ctx, network, net.JoinHostPort(ip, port))
		}
	} else {
		transport.DialContext = dialer.DialContext
	}

	return transport
}

// doRequest send request with http client to server
func (z *Zhttp) doRequest(method, rawURL string, options *ReqOptions, jar http.CookieJar) (*Response, error) {
	if options == nil {
		options = &ReqOptions{}
	}

	rawURL, err := z.buildURL(rawURL, options)
	if err != nil {
		return nil, err
	}

	req, err := z.buildRequest(method, rawURL, options)
	if err != nil {
		return nil, err
	}

	oldHost, set := z.parseIPOfDomain(req, options)

	z.addCookies(req, options)
	z.addHeaders(req, options)

	client := z.buildClient(z.options, jar)
	if options.Timeout > 0 {
		client.Timeout = options.Timeout
	}

	resp, err := client.Do(req)
	if set {
		req.URL.Host = oldHost
	}

	if err != nil {
		return nil, err
	}

	return &Response{
		RawResponse:   resp,
		StatusCode:    resp.StatusCode,
		Status:        resp.Status,
		ContentLength: resp.ContentLength,
	}, nil
}

// buildRequest build request with body and other
func (z *Zhttp) buildRequest(method, rawURL string, options *ReqOptions) (*http.Request, error) {
	ctx := context.Background()
	if options.DisableRedirect || len(options.Proxies) > 0 {
		ctx = context.WithValue(ctx, "options", options)
	}

	if options.Body == nil {
		return http.NewRequestWithContext(ctx, method, rawURL, nil)
	}

	bodyReader, contentType, err := options.Body.Reader()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, rawURL, bodyReader)
	if err != nil {
		return nil, err
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	return req, nil
}

// buildURL make url and set custom query
func (z *Zhttp) buildURL(rawURL string, options *ReqOptions) (string, error) {
	if len(options.Query) > 0 {
		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			return "", err
		}

		if parsedURL.RawQuery != "" {
			parsedURL.RawQuery += "&" + options.Query.Encode()
		} else {
			parsedURL.RawQuery = options.Query.Encode()
		}

		rawURL = parsedURL.String()
	}

	return rawURL, nil
}

// parseIPOfDomain handle custom dns resolution
func (z *Zhttp) parseIPOfDomain(req *http.Request, options *ReqOptions) (string, bool) {
	if options.IPOfDomain != "" {
		oldHost := req.URL.Host
		port := req.URL.Port()
		if port != "" {
			req.URL.Host = options.IPOfDomain + ":" + port
		} else {
			req.URL.Host = options.IPOfDomain
		}
		return oldHost, true
	}
	return "", false
}

// addHeaders handle custom headers
func (z *Zhttp) addHeaders(req *http.Request, options *ReqOptions) {
	req.Header.Set("User-Agent", "Zhttp/2.0")

	for key, value := range z.options.Headers {
		req.Header.Set(key, value)
	}

	if z.options.UserAgent != "" {
		req.Header.Set("User-Agent", z.options.UserAgent)
	}

	for key, value := range options.Headers {
		req.Header.Set(key, value)
	}

	if options.Host != "" {
		req.Host = options.Host
	}

	if options.Auth.Username != "" {
		req.SetBasicAuth(options.Auth.Username, options.Auth.Password)
	}

	if options.IsAjax {
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
	}

	if options.ContentType != "" {
		req.Header.Set("Content-Type", options.ContentType)
	}

	if options.UserAgent != "" {
		req.Header.Set("User-Agent", options.UserAgent)
	}
}

// addCookies handle custom cookies
func (z *Zhttp) addCookies(req *http.Request, options *ReqOptions) {
	for k, v := range options.Cookies {
		req.AddCookie(&http.Cookie{Name: k, Value: v})
	}
}
