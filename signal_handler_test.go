package gograce

import (
	"context"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSignalHandler(t *testing.T) {
	t.Run("without force", func(t *testing.T) {
		sh, ctx := NewSignalHandler(context.Background(), SignalHandlerOptions{
			Force: false,
		})

		require.True(t, sh.started.Load())

		go func() {
			sh.sigChan <- syscall.SIGINT
		}()

		<-ctx.Done()
		require.ErrorIs(t, ctx.Err(), context.Canceled)
	})

	t.Run("with force", func(t *testing.T) {
		var forceCalled bool
		sh, ctx := NewSignalHandler(context.Background(), SignalHandlerOptions{
			Force: true,
		})

		sh.forceFunc = func() {
			forceCalled = true
		}

		go func() {
			sh.sigChan <- syscall.SIGINT
			sh.sigChan <- syscall.SIGINT
		}()

		<-ctx.Done()
		require.ErrorIs(t, ctx.Err(), context.Canceled)
		require.True(t, forceCalled)
		require.False(t, sh.started.Load())
	})

	t.Run("cancel parent context", func(t *testing.T) {
		var forceCalled bool
		ctx, cancel := context.WithCancel(context.Background())
		sh, ctx := NewSignalHandler(ctx, SignalHandlerOptions{
			Force: true,
			ForceFunc: func() {
				forceCalled = true
			},
		})

		cancel()

		time.Sleep(time.Millisecond * 100)

		require.ErrorIs(t, ctx.Err(), context.Canceled)
		require.False(t, forceCalled)
		require.False(t, sh.started.Load())
	})

	t.Run("multiple start and close", func(t *testing.T) {
		sh, ctx := NewSignalHandler(context.Background(), SignalHandlerOptions{
			Force: false,
		})

		require.True(t, sh.started.Load())

		go func() {
			sh.sigChan <- syscall.SIGINT
		}()

		<-ctx.Done()
		require.ErrorIs(t, ctx.Err(), context.Canceled)

		sh.Start(context.Background())

		require.True(t, sh.started.Load())

		go func() {
			sh.sigChan <- syscall.SIGINT
		}()

		<-ctx.Done()
		require.ErrorIs(t, ctx.Err(), context.Canceled)

	})

	t.Run("calling close", func(t *testing.T) {
		sh, _ := NewSignalHandler(context.Background(), SignalHandlerOptions{
			Force: false,
		})

		require.True(t, sh.started.Load())

		sh.Close()

		require.False(t, sh.started.Load())
	})

	t.Run("calling close while force quiting", func(t *testing.T) {
		var forceCalled bool
		sh, _ := NewSignalHandler(context.Background(), SignalHandlerOptions{
			Force: true,
			ForceFunc: func() {
				forceCalled = true
			},
		})

		require.True(t, sh.started.Load())

		sh.sigChan <- syscall.SIGINT

		time.Sleep(100 * time.Millisecond)
		require.True(t, sh.started.Load())

		sh.Close()

		require.False(t, sh.started.Load())
		require.False(t, forceCalled)
	})

	t.Run("calling start multiple times", func(t *testing.T) {
		var (
			forceCalled bool
			wg          = sync.WaitGroup{}
		)

        wg.Add(1)
		sh, _ := NewSignalHandler(context.Background(), SignalHandlerOptions{
			Force: true,
			ForceFunc: func() {
                defer wg.Done()
				forceCalled = true
			},
		})

		sh.Start(context.Background())
		sh.Start(context.Background())
		sh.Start(context.Background())
		sh.Start(context.Background())

		require.True(t, sh.started.Load())

		sh.sigChan <- syscall.SIGINT
		sh.sigChan <- syscall.SIGINT
        wg.Wait()

		require.False(t, sh.started.Load())
		require.True(t, forceCalled)
	})
}
