package main

import (
	"context"
	"fmt"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/options"
	"net/http"
	"testing"
)

var opt *options.CollectorRunOptions

// from Go 1.13, you'll get the following error if you use flag.Parse() in init()
// https://stackoverflow.com/questions/27342973/custom-command-line-flags-in-gos-unit-tests
func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	opt = options.Parse()

	fmt.Println("attempting to run test collector for coverage data")
	killServer := newKillServer(":19999", cancel)
	go killServer.Start()
	go run("0.0.0", opt)

	<-ctx.Done()

	killServer.server.Shutdown(context.Background())
}

type killServer struct {
	server http.Server
	cancel context.CancelFunc
}

func newKillServer(addr string, cancel context.CancelFunc) *killServer {
	return &killServer{
		server: http.Server{
			Addr: addr,
		},
		cancel: cancel,
	}
}

func (s *killServer) Start() {
	s.server.Handler = s

	err := s.server.ListenAndServe()
	if err != nil {
		fmt.Println("KillServer Error:", err)
	}
}

func (s *killServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	// cancel the context
	s.cancel()
}
