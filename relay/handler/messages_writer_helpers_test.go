package controller

// minimal gin.ResponseWriter fake for the Anthropic writer tests.

import (
	"bufio"
	"bytes"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

type fakeGinWriter struct {
	buf     *bytes.Buffer
	status  int
	written bool
	size    int
}

func (w *fakeGinWriter) Header() http.Header { return http.Header{} }
func (w *fakeGinWriter) Write(data []byte) (int, error) {
	w.written = true
	w.size += len(data)
	return w.buf.Write(data)
}
func (w *fakeGinWriter) WriteString(s string) (int, error) {
	w.written = true
	w.size += len(s)
	return w.buf.WriteString(s)
}
func (w *fakeGinWriter) WriteHeader(code int) { w.status = code; w.written = true }
func (w *fakeGinWriter) WriteHeaderNow()      {}
func (w *fakeGinWriter) Status() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}
func (w *fakeGinWriter) Size() int     { return w.size }
func (w *fakeGinWriter) Written() bool { return w.written }
func (w *fakeGinWriter) Flush()        {}
func (w *fakeGinWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, nil
}
func (w *fakeGinWriter) CloseNotify() <-chan bool { ch := make(chan bool, 1); return ch }

func (w *fakeGinWriter) Pusher() (pusher http.Pusher) { return nil }

var _ gin.ResponseWriter = (*fakeGinWriter)(nil)
