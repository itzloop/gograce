package gograce

import (
	"context"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGracefulForce(t *testing.T) {
	grace := NewGracefulWithContext(context.Background(), Options{
		NoForceQuit: false,
	})

	var (
		started           bool
		ended             bool
		forceHanlerCalled bool
		wg                = sync.WaitGroup{}
	)

	grace.GoWithContext(func(ctx context.Context) error {
		started = true
		return nil
	})

	grace.GoWithContext(func(ctx context.Context) error {
        time.Sleep(100 * time.Millisecond)
		ended = true
		return nil
	})

	wg.Add(1)
	grace.sh.forceHandler = func() {
		defer wg.Done()
		forceHanlerCalled = true
	}

	go func() {
		grace.sh.sigChan <- syscall.SIGINT
		grace.sh.sigChan <- syscall.SIGINT
	}()

	wg.Wait()
	assert.True(t, forceHanlerCalled)
	assert.True(t, started)
	assert.False(t, ended)
}

func TestGracefulTimeout(t *testing.T) {
	grace := NewGracefulWithContext(context.Background(), Options{
		Timeout:     time.Millisecond,
		NoForceQuit: false,
	})

	var (
		started            bool
		ended              bool
		timeoutHandlerFunc bool
		wg                 = sync.WaitGroup{}
	)

	grace.GoWithContext(func(ctx context.Context) error {
		started = true
		return nil
	})

	grace.GoWithContext(func(ctx context.Context) error {
        time.Sleep(100 * time.Millisecond)
		ended = true
		return nil
	})

	go func() {
		grace.sh.sigChan <- syscall.SIGINT
	}()

	wg.Add(1)
	grace.th.timeoutFunc = func() {
		defer wg.Done()
		timeoutHandlerFunc = true
	}

	wg.Wait()

	assert.True(t, timeoutHandlerFunc)
	assert.True(t, started)
	assert.False(t, ended)

}

func TestGraceful(t *testing.T) {
	grace := NewGracefulWithContext(context.Background(), Options{
		Timeout:     1 * time.Second,
		NoForceQuit: false,
	})

	var (
		started bool
		ended   bool
		wg      = sync.WaitGroup{}
	)

	grace.GoWithContext(func(ctx context.Context) error {
		started = true
		return nil
	})

	wg.Add(1)
	grace.GoWithContext(func(ctx context.Context) error {
		defer wg.Done()
		ended = true
		return nil
	})

	go func() {
		grace.sh.sigChan <- syscall.SIGINT
	}()

	wg.Wait()

	assert.True(t, started)
	assert.True(t, ended)

}
