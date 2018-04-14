# zhttp

zhttp 是一个对 net/http 标准库的封装，参考了 python 中著名的 requests 库，解决了 golang 原生 http 库在高并发中 timeout 失效导致线程不能退出的的问题，很多代码及思路来自[grequests](https://github.com/levigross/grequests)

## Installation

```
go get github.com/Greyh4t/zhttp
```

## Example

如下为简单示例，更多使用方法请参考源码

```go
package main

import (
	"fmt"
	"time"

	"github.com/Greyh4t/zhttp"
)

func main() {
	//	proxy, _ := url.Parse("http://127.0.0.1:8080")

	resp, err := zhttp.Get("https://www.jd.com/?key1=value1", &zhttp.RequestOptions{
		DialTimeout:        time.Second * 5,
		RequestTimeout:     time.Second * 5,
		DisableRedirect:    true,
		InsecureSkipVerify: true,
		ContentType:        "application/x-www-form-urlencoded",
		Params: map[string]string{
			"key2": "value2",
		},
		RawCookie: "key1=value1; key2=value2",
		//		Proxies: map[string]*url.URL{
		//			"http":  proxy,
		//			"https": proxy,
		//		},
	})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(resp.StatusCode())
		fmt.Println(len(resp.String()))
		fmt.Println(resp.RawRequest())
		resp.Close()
	}
}
```


## 参考

[grequests](https://github.com/levigross/grequests)
