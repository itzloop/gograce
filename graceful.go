package gograceful

import (
	"context"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var defaultSignals = [...]os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP}

type Options struct {
	// Timeout defines how long should the program wait before forcefully exiting.
	// a zero-value indicates no timeout.
	Timeout time.Duration

	// NoForceQuit disables the force quit feature. After the first termination signal, any further signals
	// will be ignored.
	NoForceQuit bool

	// MaxGoRoutines defines how many go-routines can be started. This value is passed to SetLimit on errgroup.Group.
	// a zero-value or negative indicates no limit.
	MaxGoRoutines int

	// TODO custom signals?
	// Signals let's you overwrite graceful.defaultSignals.
	// a zero-value or an empty slice indicate no overwrite
	Signals []os.Signal
}

type Graceful struct {
	// TODO is this a good idea? [issue#22602](https://github.com/golang/go/issues/22602)
	// TODO I don't like doing a bunch of g.Go(func() error { return f(ctx) })
	// TODO instead i'd like to do g.Go(f)
	ctx context.Context
	g   *errgroup.Group
}

// NewGraceful calls NewGracefulWithContext with context.Background()
func NewGraceful(opts Options) *Graceful {
	return NewGracefulWithContext(context.Background(), opts)
}

func NewGracefulWithContext(ctx context.Context, opts Options) *Graceful {
	var (
		g        *errgroup.Group
		graceful = &Graceful{}
		signals  = defaultSignals[:]
	)

	// run signal handler
	if len(opts.Signals) != 0 {
		signals = opts.Signals
	}

	ctx = graceful.signalHandler(ctx, !opts.NoForceQuit, signals)

	if opts.Timeout != 0 {
		go graceful.timeoutHandler(ctx, opts.Timeout)
	}

	g, ctx = errgroup.WithContext(ctx)

	if opts.MaxGoRoutines != 0 {
		g.SetLimit(opts.MaxGoRoutines)
	}

	graceful.g = g
	graceful.ctx = ctx

	return graceful
}

func (grace *Graceful) signalHandler(ctx context.Context, force bool, signals []os.Signal) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, signals...)
		s := <-sig
		log.Printf("received signal '%s', gracefully quitting...", s) // TODO logging
		cancel()                                                      // instead of calling close directly we call context.CancelFunc when a signal is received

		if force {
			s = <-sig                                                     // listen again to force quit
			log.Printf("received signal '%s', forcefully quitting...", s) // TODO logging
			os.Exit(1)                                                    // forcefully terminate the program
		}
	}()

	return ctx
}

func (grace *Graceful) timeoutHandler(ctx context.Context, timeout time.Duration) {
	<-ctx.Done() // make sure we are in termination phase

	// create a timer to be able to handle timeouts
	time.AfterFunc(timeout, func() {
		log.Println("timeoutHandler: cleanup phase timeout reached, forcefully quitting...")
		os.Exit(1)
	})
}

func (grace *Graceful) GoWithContext(f func(ctx context.Context) error) {
	grace.g.Go(func() error {
		return f(grace.ctx)
	})
}
func (grace *Graceful) Go(f func() error) {
	grace.g.Go(f)
}

func (grace *Graceful) Wait() error {
	return grace.g.Wait()
}

func (grace *Graceful) FatalWait() {
	if err := grace.Wait(); err != nil {
		log.Fatalln(err.Error())
	}
}
