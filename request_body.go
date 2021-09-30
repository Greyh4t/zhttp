package zhttp

import (
	"io"
	"time"
)

type RequestBody struct {
	rc      io.ReadCloser
	timeout time.Duration
	timer   *time.Timer
}

func NewRequestBody(rc io.ReadCloser, timer *time.Timer, timeout time.Duration) *RequestBody {
	return &RequestBody{
		rc:      rc,
		timeout: timeout,
		timer:   timer,
	}
}

func (r *RequestBody) Read(p []byte) (int, error) {
	// 这是一个hack的方法
	// http请求的body在底层会通过io.Copy写入到网络连接中
	// io.Copy会循环调用Read并将数据写入网络连接
	// 因此在这里开始计时等待读取和写入超时，直到下一次计时器被重置
	r.timer.Reset(r.timeout)

	return r.rc.Read(p)
}

func (r *RequestBody) Close() error {
	// Body写入完成后，底层会关闭请求Body，因此通过Close方法在body发送完成后再次重置计时器，等待response header返回
	r.timer.Reset(r.timeout)
	return r.rc.Close()
}
