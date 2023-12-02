package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"
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

	// create a sync.WaitGroup and add 1 to it for each go-routine
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := runApplication(ctx)
		if err != nil {
			log.Printf("main: error in application: %v\n", err)
		}
	}()

	// run the timeoutHandler handler to have a hard time limit on cleanup phase
	go timeoutHandler(ctx, 30*time.Second)

	// we wait for all the go-routines
	wg.Wait()
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
	log.Println("application done.")
	return errors.New("intentional error")
}
