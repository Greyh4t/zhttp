package zhttp

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/greyh4t/dnscache"
)

var ctxOptionKey = struct{}{}

func disableRedirect(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

// buildClient make a new client
func (z *Zhttp) buildClient(httpOptions *HTTPOptions, reqOptions *ReqOptions, cookieJar http.CookieJar) *http.Client {
	client := &http.Client{
		Transport: z.transport,
		Jar:       cookieJar,
		Timeout:   httpOptions.Timeout,
	}

	if reqOptions.Timeout > 0 {
		client.Timeout = reqOptions.Timeout
	}

	if reqOptions.DisableRedirect {
		client.CheckRedirect = disableRedirect
	}

	return client
}

// createTransport create a global *http.Transport for all http client
func createTransport(options *HTTPOptions, cache *dnscache.Cache) *http.Transport {
	transport := http.DefaultTransport.(*http.Transport).Clone()

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

	transport.Proxy = func(req *http.Request) (*url.URL, error) {
		reqOptions, ok := req.Context().Value(ctxOptionKey).(*ReqOptions)
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

	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}
	if options.DialTimeout > 0 {
		dialer.Timeout = options.DialTimeout
	}
	if options.KeepAlive > 0 {
		dialer.KeepAlive = options.KeepAlive
	}

	if cache != nil {
		transport.DialContext = func(ctx context.Context, network string, address string) (net.Conn, error) {
			host, port, _ := net.SplitHostPort(address)
			ip, err := cache.FetchOneV4String(host)
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

	ctx := context.Background()
	req, err := z.buildRequest(ctx, method, rawURL, options)
	if err != nil {
		return nil, err
	}

	oldHost, set := z.parseHostIP(req, options)

	z.addCookies(req, options)
	z.addHeaders(req, options)

	client := z.buildClient(z.options, options, jar)

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
		Headers:       Headers(resp.Header),
		Body:          &ZBody{rawBody: resp.Body},
	}, nil
}

// buildRequest build request with body and other
func (z *Zhttp) buildRequest(ctx context.Context, method, rawURL string, options *ReqOptions) (*http.Request, error) {
	if len(options.Proxies) > 0 {
		ctx = context.WithValue(ctx, ctxOptionKey, options)
	}

	if options.Body == nil {
		return http.NewRequestWithContext(ctx, method, rawURL, nil)
	}

	bodyReader, contentType, err := options.Body.Content()
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

// parseHostIP handle custom dns resolution
func (z *Zhttp) parseHostIP(req *http.Request, options *ReqOptions) (string, bool) {
	if options.HostIP != "" {
		oldHost := req.URL.Host
		port := req.URL.Port()
		if port != "" {
			req.URL.Host = options.HostIP + ":" + port
		} else {
			req.URL.Host = options.HostIP
		}
		return oldHost, true
	}
	return "", false
}

// addHeaders handle custom headers
func (z *Zhttp) addHeaders(req *http.Request, options *ReqOptions) {
	if !(z.options.NoUA || options.NoUA) {
		req.Header.Set("User-Agent", "Zhttp/2.0")
	}

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
	for k, v := range z.options.Cookies {
		if _, ok := options.Cookies[k]; !ok {
			req.AddCookie(&http.Cookie{Name: k, Value: v})
		}
	}

	for k, v := range options.Cookies {
		req.AddCookie(&http.Cookie{Name: k, Value: v})
	}
}
