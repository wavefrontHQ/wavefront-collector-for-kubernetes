package main

import (
	"context"
	"fmt"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/options"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"
	"net/http"
	"os"
	"testing"
)

var collectorArgs []string

func TestMain(m *testing.M) {
	//testFS := pflag.NewFlagSet(fmt.Sprintf("%s.test", os.Args[0]), pflag.ContinueOnError)
	//testFS.StringVar(&version, "test-version", "", "set version with flag for testing because ldflags don't work properly")
	//testFS.StringVar(&commit, "test-commit", "", "set commit with flag for testing because ldflags don't work properly")
	//if err := testFS.Parse(os.Args[1:3]); err != nil {
	//	fmt.Println(err)
	//	os.Exit(2)
	//}

	version = "1.12.0"
	commit = "4930b29"

	fmt.Println(fmt.Sprintf("attempting to run test collector for coverage data with version '%s' and commit '%s'", version, commit))

	fmt.Println(fmt.Sprintf("arg stuff BEFORE shenanigans: collectorArgs '%+v' os.Args '%+v'", collectorArgs, os.Args))
	collectorArgs = os.Args[2:]
	os.Args = os.Args[:2]
	fmt.Println(fmt.Sprintf("arg stuff AFTER shenanigans: collectorArgs '%+v' os.Args '%+v'", collectorArgs, os.Args))

	os.Exit(m.Run())
}

func TestMainCoverage(t *testing.T) {
	// TODO consider making this more legit
	if collectorArgs[0] != "--daemon" {
		t.Skip("skipping collector coverage test: it appears a normal go test is being run")
	}

	ctx, cancel := context.WithCancel(context.Background())
	ks := newKillServer(":19999", cancel)
	go ks.Start()

	os.Args = append([]string{"./wavefront-collector.test"}, collectorArgs...)
	go main()

	util.SetAgentType(options.AllAgentType)

	<-ctx.Done()

	fmt.Println("context done; attempting to shut down")
	ks.server.Shutdown(context.Background())
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

	fmt.Println("receiving kill curl; attempting context cancel")
	// cancel the context
	s.cancel()
}
