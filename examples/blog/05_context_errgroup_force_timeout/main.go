package main

import (
	"context"
	"errors"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func init() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)
}

func main() {
	ctx := context.Background()

	// start signal handler
	ctx = signalHandler(ctx)

	// create a group from errgroup package
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return runApplication(ctx)
	})

	// run the timeoutHandler handler to have a hard time limit on cleanup phase
	go timeoutHandler(ctx, 30*time.Second)

	if err := group.Wait(); err != nil {
		log.Printf("one of the go-routines failed: %v", err)
		os.Exit(1)
	}
}

func signalHandler(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		s := <-sig
		log.Printf("received signal '%s', gracefully quitting...", s)
		cancel()  // instead of calling close directly we call context.CancelFunc when a signal is received
		s = <-sig // listen again to force quit
		log.Printf("received signal '%s', forcefully quitting...", s)
		os.Exit(1) // forcefully terminate the program
	}()

	return ctx
}

func timeoutHandler(ctx context.Context, timeout time.Duration) {
	<-ctx.Done() // make sure we are in termination phase

	// create a timer to be able to handle timeouts
	time.AfterFunc(timeout, func() {
		log.Println("timeoutHandler: cleanup phase timeout reached, forcefully quitting...")
		os.Exit(1)
	})
}

func runApplication(ctx context.Context) (err error) {
	defer func() {
		err = cleanupApplication()
	}()

	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			if err == context.Canceled {
				err = nil
			}

			return
		default:
			log.Println("doing stuff...")
			time.Sleep(time.Second)
		}
	}
}

func cleanupApplication() error {
	time.Sleep(5 * time.Second)
	//log.Println("application done.")
	return errors.New("intentional error")
}
