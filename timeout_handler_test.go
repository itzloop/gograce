package gograce

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTimeoutHandler(t *testing.T) {
	var (
		timeoutFuncCalled bool
		wg                = sync.WaitGroup{}
		timeoutFunc       = func() {
			defer wg.Done()
			timeoutFuncCalled = true
		}
	)

	ctx, cancel := context.WithCancel(context.Background())

	wg.Add(1)
	NewTimeoutHandler(ctx, TimeoutHandlerOptions{
		Timeout:     100 * time.Millisecond,
		TimeoutFunc: timeoutFunc,
	})

	cancel()

	wg.Wait()

	require.True(t, timeoutFuncCalled)
}
