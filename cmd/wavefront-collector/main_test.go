package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
)

// from Go 1.13, you'll get the following error if you use flag.Parse() in init()
// https://stackoverflow.com/questions/27342973/custom-command-line-flags-in-gos-unit-tests
func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())

	fmt.Println(fmt.Sprintf("Args in TestMain '%+v'", os.Args))

	// This will be so cool when I know how to make Kubernetes not destroy my Dockerfile entrypoint
	//testFS := pflag.NewFlagSet(fmt.Sprintf("%s.test", os.Args[0]), pflag.ContinueOnError)
	//testFS.StringVar(&version, "test-version", "", "set version with flag for testing because ldflags don't work properly")
	//testFS.StringVar(&commit, "test-commit", "", "set commit with flag for testing because ldflags don't work properly")
	//if err := testFS.Parse(os.Args[1:3]); err != nil {
	//	fmt.Println(err)
	//	os.Exit(2)
	//}
	//
	//newArgs := []string{os.Args[0]}
	//newArgs = append(newArgs, os.Args[3:]...)
	//os.Args = newArgs

	version = "1.12.0"
	commit = "4930b29"

	fmt.Println(fmt.Sprintf("attempting to run test collector for coverage data with version '%s' and commit '%s'", version, commit))
	killServer := newKillServer(":19999", cancel)
	go killServer.Start()
	go main()

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
