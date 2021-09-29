package zhttp

import (
	"net/url"
	"time"
)

// Auth is used to perform HTTP Basic authentication
type Auth struct {
	Username string
	Password string
}

// M is short for map[string] string
type M map[string]string

// P is short for map[string]*url.URL
type P map[string]*url.URL

// HTTPOptions is the options for zhttp.Zhttp,
// it will effective on per request
type HTTPOptions struct {
	// UserAgent allows you to set an arbitrary custom user agent
	UserAgent string

	// Cookie allows you to attach cookies to every request.
	Cookies map[string]string

	// Headers uses to set custom HTTP headers to every request
	// The header name is case-sensitive
	Headers map[string]string

	// Proxies is a map in the following format
	// *protocol* => proxy address e.g http => http://127.0.0.1:8080,
	Proxies map[string]*url.URL

	// InsecureSkipVerify is a flag that specifies if we should validate the
	// server's TLS certificate. It should be noted that Go's TLS verify mechanism
	// doesn't validate if a certificate has been revoked
	InsecureSkipVerify bool

	// RequestTimeout is the maximum amount of time a whole request(include dial / request / redirect) will wait
	RequestTimeout time.Duration

	// Timeout is the time that the client will wait between bytes sent from the server.
	Timeout time.Duration

	// DialTimeout is the maximum amount of time a dial will wait for a connect to complete
	DialTimeout time.Duration

	// TLSHandshakeTimeout specifies the maximum amount of time waiting to
	// wait for a TLS handshake. Zero means no timeout.
	TLSHandshakeTimeout time.Duration

	// KeepAlive specifies the interval between keep-alive
	// probes for an active network connection.
	// If zero, keep-alive probes are sent with a default value
	// (currently 15 seconds), if supported by the protocol and operating
	// system. Network protocols or operating systems that do
	// not support keep-alives ignore this field.
	// If negative, keep-alive probes are disabled.
	KeepAlive time.Duration

	// DisableKeepAlives, if true, disables HTTP keep-alives and
	// will only use the connection to the server for a single
	// HTTP request.
	//
	// This is unrelated to the similarly named TCP keep-alives.
	DisableKeepAlives bool

	// DisableCompression, if true, prevents the Transport from
	// requesting compression with an "Accept-Encoding: gzip"
	// request header when the Request contains no existing
	// Accept-Encoding value. If the Transport requests gzip on
	// its own and gets a gzipped response, it's transparently
	// decoded in the Response.Body. However, if the user
	// explicitly requested gzip it is not automatically
	// uncompressed.
	DisableCompression bool

	// MaxIdleConns controls the maximum number of idle (keep-alive)
	// connections across all hosts. Zero means no limit.
	MaxIdleConns int

	// MaxIdleConnsPerHost, if non-zero, controls the maximum idle
	// (keep-alive) connections to keep per-host. If zero,
	// DefaultMaxIdleConnsPerHost is used.
	MaxIdleConnsPerHost int

	// MaxConnsPerHost optionally limits the total number of
	// connections per host, including connections in the dialing,
	// active, and idle states. On limit violation, dials will block.
	//
	// Zero means no limit.
	MaxConnsPerHost int

	// IdleConnTimeout is the maximum amount of time an idle
	// (keep-alive) connection will remain idle before closing
	// itself.
	// Zero means no limit.
	IdleConnTimeout time.Duration

	// DNSCacheExpire is the timeout of dns cache , if zero, not use dns cache
	DNSCacheExpire time.Duration

	// DNSServer allows you to set an custom dns host, like 1.1.1.1:25, only effective in linux
	DNSServer string
}

// ReqOptions is the options for single request
type ReqOptions struct {
	// RequestTimeout is the maximum amount of time a whole request(include dial / request / redirect) will wait.
	// if non-zero, overwrite HTTPOptions.Timeout in current request.
	RequestTimeout time.Duration

	// Timeout is the time that the client will wait between bytes sent from the server.
	Timeout time.Duration

	// ContentType allows you to set an arbitrary custom content type
	ContentType string

	// UserAgent allows you to set an arbitrary custom user agent
	UserAgent string

	// Proxies is a map in the following format
	// *protocol* => proxy address e.g http => http://127.0.0.1:8080,
	// If setted, overwrite HTTPOptions.Proxies in current request.
	Proxies map[string]*url.URL

	// DisableRedirect will disable redirect for request
	DisableRedirect bool

	// Query will be encode to query string that may be used within a GET request
	Query url.Values

	// Body is a interface{} that will eventually convert into the the body of a POST request
	Body Body

	// Cookie allows you to attach cookies to your request.
	// Only effective in current request
	Cookies map[string]string

	// Headers uses to set custom HTTP headers to the request
	// The header name is case-sensitive
	Headers map[string]string

	// Host allows you to set an arbitrary custom host
	Host string

	// HostIP allows you to set an custom dns resolution for current request.
	// The value should be an IP.
	// When proxy usable, this value does not take effect
	HostIP string

	// Auth allows you to specify a user name and password that you wish to
	// use when requesting the URL. It will use basic HTTP authentication
	// formatting the username and password in base64.
	Auth Auth

	// IsAjax is a flag that can be set to make the request appear
	// to be generated by browser Javascript.
	IsAjax bool
}
