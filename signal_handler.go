package gograce

import (
	"context"
	"log"
	"os"
	"syscall"

	"os/signal"
)

var defaultSignals = [...]os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP}

type SignalHandler struct {
	signals []os.Signal
	force   bool
	sigChan chan os.Signal

	// THIS IS FOR TESTING PURPOSES. normally this will be just os.Exit(1)
	forceHandler func()
}

func NewSignalHandler(ctx context.Context, force bool) (*SignalHandler, context.Context) {
	return NewSignalHandlerWithSignals(ctx, force)
}

func NewSignalHandlerWithSignals(ctx context.Context, force bool, signals ...os.Signal) (*SignalHandler, context.Context) {
	if len(signals) == 0 {
		signals = defaultSignals[:]
	}

	sh := &SignalHandler{
		signals:      signals,
		force:        force,
	}

	ctx = sh.start(ctx)

	return sh, ctx
}

func (sh *SignalHandler) start(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	sh.sigChan = make(chan os.Signal, 1)
	go func() {
		signal.Notify(sh.sigChan, sh.signals...)

		defer close(sh.sigChan)

		s := <-sh.sigChan
		log.Printf("received signal '%s', gracefully quitting...", s) // TODO logging

		cancel() // instead of calling close directly we call context.CancelFunc when a signal is received

		if sh.force {
			s = <-sh.sigChan                                              // listen again to force quit
			log.Printf("received signal '%s', forcefully quitting...", s) // TODO logging

            if sh.forceHandler != nil {
                sh.forceHandler()
            } else {
                os.Exit(1)
            }
		}
	}()

	return ctx
}
