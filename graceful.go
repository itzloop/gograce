package gograce

import (
	"context"
	"log"
	"os"
	"time"

	"golang.org/x/sync/errgroup"
)

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
	sh  *SignalHandler
	th  *TimeoutHandler
}

// NewGraceful calls NewGracefulWithContext with context.Background()
func NewGraceful(opts Options) *Graceful {
	return NewGracefulWithContext(context.Background(), opts)
}

// NewGracefulWithContext will create a SignalHandler and a TimeoutHandler
// which are started automatically.
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

	// Create signal handler
	graceful.sh, ctx = NewSignalHandler(ctx, SignalHandlerOptions{
		Force:   !opts.NoForceQuit,
		Signals: signals,
	})

	if opts.Timeout != 0 {
		graceful.th = NewTimeoutHandler(ctx, TimeoutHandlerOptions{
			Timeout: opts.Timeout,
		})
	}

	g, ctx = errgroup.WithContext(ctx)

	if opts.MaxGoRoutines != 0 {
		g.SetLimit(opts.MaxGoRoutines)
	}

	graceful.g = g
	graceful.ctx = ctx

	return graceful
}

// GoWithContext is convenient wrapper for (*errgroup.Group).Go that
// accepts a functions that takes a context as input instead of not
// having any input.
func (grace *Graceful) GoWithContext(f func(ctx context.Context) error) {
	grace.g.Go(func() error {
		return f(grace.ctx)
	})
}

// Go calls (*errgroup.Group).Go() internally
func (grace *Graceful) Go(f func() error) {
	grace.g.Go(f)
}

// Wait calls (*errgroup.Group).Wait() and returns the error
func (grace *Graceful) Wait() error {
	return grace.g.Wait()
}

// FatalWait calls Wait but log.Fatals when an error is received
func (grace *Graceful) FatalWait() {
	if err := grace.Wait(); err != nil {
		log.Fatalln(err.Error())
	}
}
