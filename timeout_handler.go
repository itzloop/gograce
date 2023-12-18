package gograce

import (
	"context"
	"log"
	"os"
	"time"
)

type TimeoutFunc func()

type TimeoutHandler struct {
	timeout time.Duration

	timeoutFunc TimeoutFunc
}

func NewTimeoutHandler(ctx context.Context, timeout time.Duration) *TimeoutHandler {
    return NewTimeoutHandlerWithTimeoutFunc(ctx, timeout, defaultTimeoutFunc)
}

func NewTimeoutHandlerWithTimeoutFunc(ctx context.Context, timeout time.Duration, timeoutFunc TimeoutFunc) *TimeoutHandler {
	th := &TimeoutHandler{
		timeout:     timeout,
		timeoutFunc: timeoutFunc,
	}

	go th.start(ctx)

	return th
}

func (th *TimeoutHandler) start(ctx context.Context) {
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
