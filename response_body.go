package zhttp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"
)

type ResponseBody struct {
	rc      io.ReadCloser
	timeout time.Duration
	timer   *time.Timer
}

func NewResponseBody(body io.ReadCloser, timer *time.Timer, timeout time.Duration) *ResponseBody {
	return &ResponseBody{
		rc:      body,
		timeout: timeout,
		timer:   timer,
	}
}

func (r *ResponseBody) Read(p []byte) (int, error) {
	r.timer.Reset(r.timeout)
	n, err := r.rc.Read(p)
	r.timer.Stop()

	if errors.Is(err, context.Canceled) {
		err = fmt.Errorf("%w (timeout exceeded while read body)", err)
	}

	return n, err
}

func (r *ResponseBody) Close() error {
	return r.rc.Close()
}
