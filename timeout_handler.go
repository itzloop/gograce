package gograce

import (
	"context"
	"log"
	"os"
	"time"
)

type TimeoutHandler struct {
	timeout time.Duration

	// THIS IS FOR TESTING PURPOSES. normally this will be just os.Exit(1)
    timeoutFunc func()
}

func NewTimeoutHandler(ctx context.Context, timeout time.Duration) *TimeoutHandler {
	th := &TimeoutHandler{
		timeout: timeout,
	}

	go th.start(ctx)

	return th
}

func (th *TimeoutHandler) start(ctx context.Context) {
	<-ctx.Done() // make sure we are in termination phase

	// create a timer to be able to handle timeouts
	time.AfterFunc(th.timeout, func() {
		log.Println("timeoutHandler: cleanup phase timeout reached, forcefully quitting...")
        if th.timeoutFunc != nil {
            th.timeoutFunc()
        } else {
            os.Exit(1)
        }
	})
}
