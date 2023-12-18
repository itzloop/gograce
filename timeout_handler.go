package gograce

import (
	"context"
	"log"
	"os"
	"time"
)

// TimeoutFunc will be called when the the
type TimeoutFunc func()

// TimeoutHandlerOptions
type TimeoutHandlerOptions struct {
	// Timeout is the value that is passed to time.AfterFunc.
	Timeout time.Duration

	// TimeoutFunc is the function that is passed to time.AfterFunc.
	TimeoutFunc TimeoutFunc
}

// TimeoutHandler will set a hard limit for graceful shutdown. If that limit
// is reached the program will call timeoutFunc. defaultTimeoutFunc is os.Exit(1)
// so this will terminate the application.
type TimeoutHandler struct {
	timeout time.Duration

	timeoutFunc TimeoutFunc
}

// NewTimeoutHandler
func NewTimeoutHandler(ctx context.Context, opts TimeoutHandlerOptions) *TimeoutHandler {
	if opts.TimeoutFunc == nil {
		opts.TimeoutFunc = defaultTimeoutFunc
	}

	th := &TimeoutHandler{
		timeout:     opts.Timeout,
		timeoutFunc: opts.TimeoutFunc,
	}

	go th.Start(ctx)

	return th
}

func (th *TimeoutHandler) Start(ctx context.Context) {
	<-ctx.Done() // make sure we are in termination phase

	// create a timer to be able to handle timeouts
	time.AfterFunc(th.timeout, func() {
		log.Println("timeoutHandler: cleanup phase timeout reached, forcefully quitting...")
		th.timeoutFunc()
	})
}

func defaultTimeoutFunc() {
	os.Exit(1)
}
