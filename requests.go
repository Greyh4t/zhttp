package zhttp

import (
	"context"
	"crypto/tls"
	"golang.org/x/net/publicsuffix"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

// buildClient make a new client
func (z *Zhttp) buildClient(options *HttpOptions, cookieJar http.CookieJar) *http.Client {
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
func (z *Zhttp) createTransport(options *HttpOptions) *http.Transport {
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
func (z *Zhttp) doRequest(method, rawUrl string, options *ReqOptions, jar http.CookieJar) (*Response, error) {
	if options == nil {
		options = &ReqOptions{}
	}

	rawUrl, err := z.buildURL(rawUrl, options)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	if options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}

	req, err := z.buildRequest(ctx, method, rawUrl, options)
	if err != nil {
		return nil, err
	}

	oldHost, set := z.parseHosts(req, options)

	z.addHeaders(req, options)
	z.addCookies(req, options)

	client := z.buildClient(z.options, jar)
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
func (z *Zhttp) buildRequest(ctx context.Context, method, rawUrl string, options *ReqOptions) (*http.Request, error) {
	if options.DisableRedirect || len(options.Proxies) > 0 {
		ctx = context.WithValue(ctx, "options", options)
	}

	if options.Body == nil {
		return http.NewRequestWithContext(ctx, method, rawUrl, nil)
	}

	if body, ok := options.Body.(*MultipartBody); ok {
		body.withContext(ctx)
	}

	bodyReader, contentType, err := options.Body.Reader()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, rawUrl, bodyReader)
	if err != nil {
		return nil, err
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	return req, nil
}

// buildURL make url and set custom query
func (z *Zhttp) buildURL(urlStr string, options *ReqOptions) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	if options.Query.Raw != "" {
		parsedURL.RawQuery = options.Query.Raw
	} else {
		if len(options.Query.Pairs) > 0 {
			query := parsedURL.Query()
			for key, value := range options.Query.Pairs {
				query.Set(key, value)
			}
			parsedURL.RawQuery = query.Encode()
		}
	}

	return parsedURL.String(), nil
}

// parseHosts handle custom dns resolve value
func (z *Zhttp) parseHosts(req *http.Request, options *ReqOptions) (string, bool) {
	if options.Hosts != "" {
		oldHost := req.URL.Host
		port := req.URL.Port()
		if port != "" {
			req.URL.Host = options.Hosts + ":" + port
		} else {
			req.URL.Host = options.Hosts
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

// addCookies handle custom cookie
func (z *Zhttp) addCookies(req *http.Request, options *ReqOptions) {
	if options.Cookie.Raw != "" {
		req.Header.Set("Cookie", options.Cookie.Raw)
	} else {
		for k, v := range options.Cookie.Pairs {
			req.AddCookie(&http.Cookie{Name: k, Value: v})
		}
	}
}
