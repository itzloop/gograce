package gograce

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeoutHandler(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	th := NewTimeoutHandler(ctx, 10*time.Microsecond)

	var (
		timeoutFuncCalled bool
        wg = sync.WaitGroup{}
	)

    wg.Add(1)
	th.timeoutFunc = func() {
        defer wg.Done()
		timeoutFuncCalled = true
	}

	cancel()
    wg.Wait()
	assert.True(t, timeoutFuncCalled)
}
