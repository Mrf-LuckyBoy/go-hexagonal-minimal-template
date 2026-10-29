package lifecycle

import (
	"context"
	"errors"
	"net/http"
	"time"
)

var _ Server = (*HTTPServer)(nil)

// HTTPServer adapts any http.Handler to Server. This covers net/http,
// Gin, Echo, Chi and every other router that implements http.Handler.
type HTTPServer struct {
	srv *http.Server
}

// NewHTTPServer wraps h in an *http.Server with safe default timeouts.
// Use NewHTTPServerFrom for full control over the http.Server settings.
func NewHTTPServer(addr string, h http.Handler, timeout time.Duration) *HTTPServer {
	return NewHTTPServerFrom(&http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       timeout,
		WriteTimeout:      timeout,
		IdleTimeout:       2 * timeout,
	})
}

// NewHTTPServerFrom wraps an already configured *http.Server.
func NewHTTPServerFrom(srv *http.Server) *HTTPServer {
	return &HTTPServer{srv: srv}
}

func (s *HTTPServer) Start(context.Context) error {
	if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
