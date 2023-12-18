package gograce

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	NewTimeoutHandlerWithTimeoutFunc(ctx, 10*time.Microsecond, timeoutFunc)

	wg.Add(1)

	cancel()

	wg.Wait()

	assert.True(t, timeoutFuncCalled)
}
