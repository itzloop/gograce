package gograce

import (
	"context"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)



func TestSignalHandlerWithoutForce(t *testing.T) {
    sh, ctx := NewSignalHandler(context.Background(), false)

    go func() {
        sh.sigChan <- syscall.SIGINT
    }()

    <-ctx.Done()
    assert.ErrorIs(t, ctx.Err(), context.Canceled)
}

func TestSignalHandlerWithForce(t *testing.T) {
    var forceCalled bool
    sh, ctx := NewSignalHandler(context.Background(), true)

    sh.forceHandler = func() {
        forceCalled = true
    }

    go func() {
        sh.sigChan <- syscall.SIGINT
        sh.sigChan <- syscall.SIGINT
    }()

    <-ctx.Done()
    assert.ErrorIs(t, ctx.Err(), context.Canceled)
    assert.True(t, forceCalled)
}
