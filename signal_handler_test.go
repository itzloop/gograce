package gograce

import (
	"context"
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
}
