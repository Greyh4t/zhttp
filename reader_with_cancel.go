package zhttp

import (
	"context"
	"fmt"
	"io"
	"time"
)

type ReaderWithCancel struct {
	rc      io.ReadCloser
	cancel  context.CancelFunc
	timeout time.Duration
	timer   *time.Timer
}

func (r *ReaderWithCancel) readWithTimeout(p []byte) (int, error) {
	if r.timer != nil {
		r.timer.Reset(r.timeout)
	} else {
		r.timer = time.AfterFunc(r.timeout, r.cancel)
	}

	n, err := r.rc.Read(p)

	r.timer.Stop()

	if err == context.Canceled {
		err = fmt.Errorf("%w (timeout exceeded while read body)", err)
	}

	return n, err
}

func (r *ReaderWithCancel) Read(p []byte) (n int, err error) {
	if r.timeout > 0 {
		return r.readWithTimeout(p)
	}

	return r.rc.Read(p)
}

func (r *ReaderWithCancel) Close() error {
	r.cancel()
	return r.rc.Close()
}
