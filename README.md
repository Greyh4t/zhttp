# zhttp

zhttp 是一个对 net/http 标准库的封装，参考了 python 中著名的 requests 库，支持标准库中的连接池，支持dns缓存，支持流式文件上传，支持多种body格式，很多代码及思路来自[grequests](https://github.com/levigross/grequests)

[![GoDoc](https://godoc.org/github.com/greyh4t/zhttp?status.svg)](https://godoc.org/github.com/greyh4t/zhttp)

## Installation

```
go get github.com/greyh4t/zhttp
```

## Usage

#### 直接使用默认client
```go
import "github.com/greyh4t/zhttp"

func main() {
	resp, err := zhttp.Get("http://www.example.com/", nil)
	if err != nil {
		return
	}
	resp.Close()
}
```

#### 更改默认client配置
```go
import "github.com/greyh4t/zhttp"

func main() {
	zhttp.InitDefaultClient(&zhttp.HTTPOptions{
		Proxies: zhttp.MustProxy(zhttp.M{
			"http":  "http://127.0.0.1:8080",
			"https": "http://127.0.0.1:8080",
		}),
	})

	resp, err := zhttp.Get("http://www.example.com/", nil)
	if err != nil {
		return
	}
	resp.Close()
}
```

#### 创建独立的client使用
```go
import "github.com/greyh4t/zhttp"

func main() {
	z := zhttp.New(&zhttp.HTTPOptions{
		Proxies: zhttp.MustProxy(zhttp.M{
			"http":  "http://127.0.0.1:8080",
			"https": "http://127.0.0.1:8080",
		}),
	})

	resp, err := z.Get("http://www.example.com/", nil)
	if err != nil {
		return
	}
	resp.Close()
}
```

## Example

如下为简单示例，更多使用方法请参考godoc

```go
package main

import (
	"log"
	"net/url"
	"os"
	"time"

	"github.com/greyh4t/zhttp"
	"github.com/greyh4t/zhttp/tools"
)

func main() {
	z := zhttp.New(&zhttp.HTTPOptions{
		UserAgent: "global-useragent",
		Headers: zhttp.M{
			"globalheader1": "value1",
			"globalheader2": "value2",
		},
		DNSCacheExpire:      time.Minute * 10,
		DNSServer:           "8.8.8.8:25",
		InsecureSkipVerify:  true,
		DialTimeout:         time.Second * 5,
		TLSHandshakeTimeout: time.Second * 5,
		KeepAlive:           time.Second * 10,
		MaxIdleConns:        10,
	})

	// 请求1
	resp, err := z.Get("http://www.example.com/", nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(resp.StatusCode, resp.Status, resp.ContentLength)
	log.Println(resp.Headers.String())
	log.Println(resp.Cookies().String())
	log.Println(resp.RawHTTPRequest())
	log.Println(resp.Body.String())
	resp.Close()

	// 请求2 post表单
	resp, err = z.Post("http://www.example.com/?query1=value3", &zhttp.ReqOptions{
		DisableRedirect: true,
		Timeout:         time.Second * 10,
		Proxies: zhttp.MustProxy(zhttp.M{
			"http":  "http://127.0.0.1:8080",
			"https": "http://127.0.0.1:8080",
		}),
		Headers: zhttp.M{
			"header1": "value1",
			"header2": "value2",
		},
		Cookies: zhttp.M{
			"k1": "v1",
			"k2": "v2",
		},
		Body: zhttp.Form(zhttp.M{
			"key1": "value1",
			"key2": "value2",
		}),
		Query: url.Values{
			"query1": {"value1"},
			"query2": {"value2"},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	body := resp.Body.Bytes()
	if resp.Body.Err != nil {
		log.Fatal(resp.Body.Err)
	}
	resp.Close()
	log.Println(body)

	// 请求3 post表单
	resp, err = z.Post("http://www.example.com/?query1=value1&query2=value2", &zhttp.ReqOptions{
		Body: zhttp.FormString(`fk1=fv1&fk2=fv2`),
		Headers: zhttp.M{
			"Cookie": "k1=v1; k2=v2",
		},
		UserAgent: "zhttp-ua-test",
	})
	if err != nil {
		log.Fatal(err)
	}
	resp.Close()

	// 请求4 post json
	resp, err = z.Post("http://www.example.com/", &zhttp.ReqOptions{
		Body: zhttp.JSONString(`{"jk1":"jv","jk2":2}`),
		Headers: zhttp.M{
			"Cookie": "k1=v1; k2=v2",
		},
		UserAgent: "zhttp-ua-test",
		IsAjax:    true,
	})
	if err != nil {
		log.Fatal(err)
	}
	resp.Close()

	// 请求5 文件上传
	f, err := os.Open("test.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	resp, err = z.Post("http://www.example.com/", &zhttp.ReqOptions{
		Body:        zhttp.Reader(f),
		ContentType: "text/plain",
		Headers: zhttp.M{
			"h1": "v1",
			"h2": "v2",
		},
		Auth: zhttp.Auth{
			Username: "username",
			Password: "password",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	resp.Close()

	// 请求6 文件上传
	file1, err := tools.FileFromDisk("file1.txt")
	if err != nil {
		log.Fatal(err)
	}

	file2, err := tools.FileFromDisk("file2.txt")
	if err != nil {
		log.Fatal(err)
	}

	resp, err = z.Post("http://www.example.com/", &zhttp.ReqOptions{
		Body: zhttp.Multipart([]*zhttp.File{file1, file2}, zhttp.M{
			"field1": "value1",
			"field2": "value2",
		}),
		Host: "file.example.com",
	})
	if err != nil {
		log.Fatal(err)
	}
	resp.Close()

	// 请求7 session的使用
	s := z.NewSession()
	resp, err = s.Post("http://www.example.com/login", &zhttp.ReqOptions{
		Body: zhttp.Form(zhttp.M{
			"username": "username",
			"password": "password",
		}),
		Timeout: time.Second * 10,
	})
	if err != nil {
		log.Fatal(err)
	}
	resp.Close()

	resp, err = s.Get("http://www.example.com/userinfo", nil)
	if err != nil {
		log.Fatal(err)
	}
	resp.Close()
}
```
