package main

import (
	"context"
	"fmt"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/options"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"
	"net/http"
	"os"
	"strings"
	"testing"
)

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

	os.Exit(m.Run())
}

func TestMainCoverage(t *testing.T) {
	collectorCoverageArgs := os.Getenv("COLLECTOR_COVERAGE_ARGS")
	if len(collectorCoverageArgs) == 0 {
		t.Skip("skipping coverage test to run in unit test mode")
	}

	collectorArgs := strings.Split(collectorCoverageArgs, " ")

	fmt.Println(fmt.Sprintf("collectorCoverageArgs '%s'", collectorCoverageArgs))
	fmt.Println(fmt.Sprintf("collectorArgs '%+v'", collectorArgs))

	ctx, cancel := context.WithCancel(context.Background())
	ks := newKillServer(":19999", cancel)
	go ks.Start()

	fmt.Println(fmt.Sprintf("collectorArgs from env: %+v", collectorArgs))
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
