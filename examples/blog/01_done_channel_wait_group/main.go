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
	// create the channel and pass it to application
	done := make(chan struct{})

	// create a sync.WaitGroup and add 1 to it for each go-routine
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		runApplication(done)
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig
	log.Println("received signal:", s)
	close(done) // we close the channel when a signal is received

	// we wait for all the go-routines
	wg.Wait()
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
