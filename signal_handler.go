package gograce

import (
	"context"
	"log"
	"os"
	"sync/atomic"
	"syscall"

	"os/signal"
)

type ForceFunc func()

var (
	defaultSignals = [...]os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP}
)

// SignalHandlerOptions
type SignalHandlerOptions struct {
	// Force enables quiting forcefully (by sending one of the Signals twice)
	// when graceful shutdown is in progress
	Force bool

	// Signals overwrites the defaultSignals.
	Signals []os.Signal

	// ForceFunc is called when Force = true and one of the Signals is sent twice.
	// If ForceFunc is nil, defaultForceFunc will be used which is os.Exit(1).
	ForceFunc ForceFunc
}

// A SignalHandler listens for signals and handles graceful and forceful shutdown
// When a signal has been sent twice SignalHandler will call forceFunc. Default
// forceFunc is os.Exit(1) so application will terminate.
type SignalHandler struct {
	signals []os.Signal
	force   bool
	sigChan chan os.Signal

	forceFunc ForceFunc

	started atomic.Bool
}

// NewSignalHandler will create a signal handler based on the desired opts given.
// It will then start the signal handler as well.
func NewSignalHandler(ctx context.Context, opts SignalHandlerOptions) (*SignalHandler, context.Context) {
	if len(opts.Signals) == 0 {
		opts.Signals = defaultSignals[:]
	}

	if opts.ForceFunc == nil {
		opts.ForceFunc = defaultForceFunc
	}

	sh := &SignalHandler{
		signals:   opts.Signals,
		force:     opts.Force,
		started:   atomic.Bool{},
		forceFunc: opts.ForceFunc,
	}

	ctx = sh.Start(ctx)

	return sh, ctx
}

// Start will start the SignalHandler by listening to signals and parent context.
// If at anypoint parent context gets canceled, Start will return. It is safe
// but useless to call Start from multiple go-routines because it will start it
// the first and you have to Close it first to be able to Start it again.
func (s *SignalHandler) Start(ctx context.Context) context.Context {
	if s.started.Swap(true) {
		return ctx // TODO should this be nil or not?
	}

	// at any point we need to stop execution when the
	// parent context gets canceled so we make a copy
	// of it.
	parentCtx := ctx
	ctx, cancel := context.WithCancel(ctx)
	s.sigChan = make(chan os.Signal, 1)

	go func() {
		signal.Notify(s.sigChan, s.signals...)

		defer s.Close()
		var (
			sig os.Signal
			ok  bool
		)

		select {
		case sig, ok = <-s.sigChan:
			if !ok {
				log.Println("signal channel closed quiting...")
				return
			}
			log.Printf("received signal '%s', gracefully quitting...\n", sig)
			cancel()
		case <-parentCtx.Done():
			log.Printf("parent context canceled\n")
			return
		}

		if s.force {
			select {
			case sig, ok = <-s.sigChan:
				if !ok {
					log.Println("signal channel closed quiting...")
					return
				}

				log.Printf("received signal '%s', forcefully quitting...\n", sig)
				cancel()
			case <-parentCtx.Done():
				log.Printf("parent context canceled, while waiting for second signal\n")
				return
			}

			s.Close()

			s.forceFunc()
			return
		}

		s.Close()
	}()

	return ctx
}

// Close closes sigChan. Calls to close only work when SignalHandler has been started
// and other wise it has no effect. It is also safe to call it from multiple go-routines.
func (sh *SignalHandler) Close() {
	if sh.started.Swap(false) {
		return
	}

	close(sh.sigChan)
}

func defaultForceFunc() {
	os.Exit(1)
}
