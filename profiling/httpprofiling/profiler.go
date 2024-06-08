package httpprofiling

import (
	"context"
	"fmt"
	"net/http"
	npprof "net/http/pprof"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"time"

	"github.com/diki-haryadi/govega/httputil"
	"github.com/diki-haryadi/govega/log"
	"github.com/diki-haryadi/govega/router"
)

const (
	defaultReadTimeout  = 5 * time.Minute
	defaultWriteTimeout = 5 * time.Minute
	defaultPort         = 8432

	stopped      uint32 = 0
	started      uint32 = 1
	shuttingdown uint32 = 2
)

type (
	Profiler struct {
		server *http.Server
		state  uint32
		lock   sync.Mutex
	}

	config struct {
		ReadTimeout  time.Duration
		WriteTimeout time.Duration
		Port         int
		Router       *router.MyRouter
		ManualStart  bool
	}

	Option func(c *config)
)

// InitProfiler, init http profiler and automatically start the http profiler by default
// To prevent automatically start the http profiler server use `WithManualStart(true)`
func InitProfiler(opts ...Option) (*Profiler, error) {
	conf := config{
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
		Port:         defaultPort,
		ManualStart:  false,
	}

	for _, o := range opts {
		o(&conf)
	}

	r := conf.Router
	if r == nil {
		r = router.New(&router.Options{
			Timeout: int(conf.WriteTimeout.Seconds()),
		})
	}

	r.Group("/debug/pprof", func(pr *router.MyRouter) {
		pr.Handler("/", http.MethodGet, http.HandlerFunc(npprof.Index))

		profiles := pprof.Profiles()
		for _, profile := range profiles {
			profileName := profile.Name()
			pr.Handler(fmt.Sprintf("/%s", profileName), http.MethodGet, npprof.Handler(profileName))
		}

		pr.Handler("/cmdline", http.MethodGet, http.HandlerFunc(npprof.Cmdline))
		pr.Handler("/profile", http.MethodGet, http.HandlerFunc(npprof.Profile))
		pr.Handler("/symbol", http.MethodGet, http.HandlerFunc(npprof.Symbol))
		pr.Handler("/symbol", http.MethodPost, http.HandlerFunc(npprof.Symbol))
		pr.Handler("/trace", http.MethodGet, http.HandlerFunc(npprof.Trace))
	})

	profiler := Profiler{
		server: &http.Server{
			Handler:      r,
			ReadTimeout:  conf.ReadTimeout,
			WriteTimeout: conf.WriteTimeout,
			Addr:         fmt.Sprintf(":%d", conf.Port),
		},
	}

	if conf.ManualStart {
		return &profiler, nil
	}

	if err := profiler.Start(); err != nil {
		return nil, fmt.Errorf("failed to start http profiler: %w", err)
	}

	return &profiler, nil
}

// Start, start the http profiler server
// this function will not do anything if already started
func (p *Profiler) Start() error {
	if atomic.LoadUint32(&p.state) == started {
		return nil
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	if atomic.LoadUint32(&p.state) == started {
		return nil
	}

	log.Printf("starting http profiler on %s\n", p.server.Addr)

	listener, err := httputil.Listen(p.server.Addr)
	if err != nil {
		return fmt.Errorf("failed to open http profiler listener on %s: %w", p.server.Addr, err)
	}

	atomic.StoreUint32(&p.state, started)

	go func() {
		defer atomic.StoreUint32(&p.state, stopped)

		if err := p.server.Serve(listener); err != nil {
			if atomic.LoadUint32(&p.state) == shuttingdown {
				log.WithError(err).Warnln("serve return error on shutting down")
				return
			}

			log.WithError(err).Errorln("failed on serving http profiler")
		}
	}()

	return nil
}

// Stop, stop the http profiler server
// this function will not do anything if already stopped
func (p *Profiler) Stop(ctx context.Context) error {
	if atomic.LoadUint32(&p.state) == stopped {
		return nil
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	if atomic.LoadUint32(&p.state) == stopped {
		return nil
	}

	atomic.StoreUint32(&p.state, shuttingdown)

	if err := p.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed on shutting down http profiler: %w", err)
	}

	return nil
}
