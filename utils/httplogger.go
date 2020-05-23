package utils

import (
	"bufio"
	"fmt"
	"go.uber.org/zap"
	"net"
	"net/http"
	"time"
)

func LogHandler(tpe string, nextHandler http.Handler) (http.Handler, error) {
	logger, err := zap.NewProduction(zap.WithCaller(false))
	if err != nil {
		return nextHandler, err
	}

	logger = logger.With(zap.String("type", tpe))
	handle := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lw := &loggedWriter{ResponseWriter: w, status: 200}
		nextHandler.ServeHTTP(lw, r)

		token := r.Header.Get("X-User-Key")
		if token == "" {
			tok, err := r.Cookie("api_token")
			if err == nil {
				token = tok.Value
			}
		}

		logger.Info("http",
			zap.String("method", r.Method),
			zap.String("url", r.RequestURI),
			zap.String("user_agent", r.UserAgent()),
			zap.String("api_key", token),
			zap.String("ip", r.Header.Get("X-Forwarded-For")),
			zap.Int64("request_size", r.ContentLength),
			zap.Int("status", lw.status),
			zap.Int("reply_size", lw.size),
			zap.Int64("duration", time.Since(start).Microseconds()),
		)
	}

	return http.HandlerFunc(handle), nil
}

type loggedWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (lw *loggedWriter) WriteHeader(status int) {
	lw.status = status
	lw.ResponseWriter.WriteHeader(status)
}

func (lw *loggedWriter) Write(b []byte) (int, error) {
	size, err := lw.ResponseWriter.Write(b)
	lw.size += size
	return size, err
}

func (lw *loggedWriter) Flush() {
	f, ok := lw.ResponseWriter.(http.Flusher)
	if ok {
		f.Flush()
	}
}

func (lw *loggedWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := lw.ResponseWriter.(http.Hijacker)
	if ok {
		return hj.Hijack()
	}

	return nil, nil, fmt.Errorf("ResponseWriter does not implement the Hijacker interface")
}
