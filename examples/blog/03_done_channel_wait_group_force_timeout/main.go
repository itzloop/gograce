package main

import (
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
	// start signal handler
	done := signalHandler()

	// create a sync.WaitGroup and add 1 to it for each go-routine
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		runApplication(done)
	}()

	// run the timeoutHandler handler to have a hard time limit on cleanup phase
	go timeoutHandler(done, 2*time.Second)

	// we wait for all the go-routines
	wg.Wait()
}

func signalHandler() <-chan struct{} {
	// create the channel and pass it to application
	done := make(chan struct{})
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		s := <-sig
		log.Printf("received signal '%s', gracefully quitting...", s)
		close(done) // we close the channel when a signal is received
		s = <-sig   // listen again to force quit
		log.Printf("received signal '%s', forcefully quitting...", s)
		os.Exit(1) // forcefully terminate the program
	}()

	return done
}

func timeoutHandler(done <-chan struct{}, timeout time.Duration) {
	<-done // make sure we are in termination phase

	// create a timer to be able to handle timeouts
	time.AfterFunc(timeout, func() {
		log.Println("timeoutHandler: cleanup phase timeout reached, forcefully quitting...")
		os.Exit(1)
	})
}

func runApplication(done <-chan struct{}) {
	defer cleanupApplication()

	for {
		select {
		case _, ok := <-done:
			if !ok {
				// exit from application when the channel is closed
				return
			}
		default:
			log.Println("doing stuff...")
			time.Sleep(time.Second)
		}
	}
}

func cleanupApplication() {
	time.Sleep(5 * time.Second)
	log.Println("application done.")
}
