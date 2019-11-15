# zhttp

zhttp 是一个对 net/http 标准库的封装，参考了 python 中著名的 requests 库，支持标准库中的连接池，支持dns缓存，支持流式文件上传，支持多种body格式，很多代码及思路来自[grequests](https://github.com/levigross/grequests)

[![GoDoc](https://godoc.org/github.com/greyh4t/zhttp?status.svg)](https://godoc.org/github.com/greyh4t/zhttp)

## Installation

```
go get github.com/Greyh4t/zhttp
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
	"github.com/Greyh4t/zhttp"
)

func main() {
	z := zhttp.New(&zhttp.HttpOptions{
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
	log.Println(resp.RawHeaders())
	log.Println(resp.CookiesMap())
	log.Println(resp.RawRequest())
	log.Println(resp.String())
	resp.Close()

	// 请求2 post表单
	resp, err = z.Post("http://www.example.com/", &zhttp.ReqOptions{
		DisableRedirect: true,
		Timeout:         time.Second * 10,
		Proxies: map[string]*url.URL{
			"http":  zhttp.MustProxy("http://127.0.0.1:8080"),
			"https": zhttp.MustProxy("http://127.0.0.1:8080"),
		},
		Body: zhttp.Form(map[string]string{
			"key1": "value1",
			"key2": "value2",
		}),
		Cookie: zhttp.PairsCookie(map[string]string{
			"k1": "v1",
			"k2": "v2",
		}),
		Query: zhttp.PairsQuery(map[string]string{
			"query1": "value1",
			"query2": "value2",
		}),
	})
	if err != nil {
		log.Fatal(err)
	}
	resp.Close()

	// 请求3 post表单
	resp, err = z.Post("http://www.example.com/", &zhttp.ReqOptions{
		Query:     zhttp.RawQuery("query1=value1&query2=value2"),
		Body:      zhttp.RawForm(`fk1=fv1&fk2=fv2`),
		Cookie:    zhttp.RawCookie("k1=v1; k2=v2"),
		UserAgent: "zhttp-ua-test",
	})
	if err != nil {
		log.Fatal(err)
	}
	resp.Close()

	// 请求4 post json
	resp, err = z.Post("http://www.example.com/", &zhttp.ReqOptions{
		Body:      zhttp.RawJSON(`{"jk1":"jv","jk2":2}`),
		Cookie:    zhttp.RawCookie("k1=v1; k2=v2"),
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

	resp, err = z.Post("http://www.example.com/", &zhttp.ReqOptions{
		Body:        zhttp.RawReader(f),
		ContentType: "text/plain",
		Headers: map[string]string{
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
	file1, err := zhttp.FileFromDisk("file1.txt")
	if err != nil {
		log.Fatal(err)
	}

	file2, err := zhttp.FileFromDisk("file2.txt")
	if err != nil {
		log.Fatal(err)
	}

	resp, err = z.Post("http://www.example.com/", &zhttp.ReqOptions{
		Body: zhttp.Multipart([]*zhttp.File{file1, file2}, map[string]string{
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
		Body: zhttp.Form(map[string]string{
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