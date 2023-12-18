package main

import (
	"context"
	"errors"
	"github.com/itzloop/gograce"
	"log"
	"net"
	"net/http"
	"time"
)

func main() {
	// create a grace instance
	grace := gograce.NewGraceful(gograce.Options{
		Timeout:       15 * time.Second, // wait 15 seconds and forcefully terminate the application
		NoForceQuit:   false,            // by pressing Ctrl+C twice, app will terminate immediately
		MaxGoRoutines: 0,                // set no limit for the number of go-routines running at the same time
		Signals:       nil,              // use the defaultSignals in signal.Notify.
	})

	// create a simple http server
	exampleHTTPServer := NewExampleHTTPServer(":8000")

	// add start and close operations to grace instance
	grace.GoWithContext(exampleHTTPServer.start)
	grace.GoWithContext(exampleHTTPServer.close)

	// wait for all go-routines or the cancel signal and
	// if any error is encountered, call log.Fatal
	grace.FatalWait()
}

type ExampleHTTPServer struct {
	// addr is the listen address for the http.Server
	addr string

	// an instance of http.Server that is used to server request
	httpServer *http.Server
}

// NewExampleHTTPServer creates an instance of ExampleHTTPServer
func NewExampleHTTPServer(addr string) *ExampleHTTPServer {
	mux := http.NewServeMux()
	mux.HandleFunc("/log-running-job", func(writer http.ResponseWriter, request *http.Request) {
		time.Sleep(10 * time.Second)
		if _, err := writer.Write([]byte("done")); err != nil {
			log.Printf("log-running-job: failed to write response: %v", err)
		}
	})
	mux.HandleFunc("/hello", func(writer http.ResponseWriter, request *http.Request) {
		if _, err := writer.Write([]byte("hi")); err != nil {
			log.Printf("log-running-job: failed to write response: %v", err)
		}
	})

	server := http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return &ExampleHTTPServer{
		addr:       addr,
		httpServer: &server,
	}
}

// start starts the httpServer and sets the http.Server.BaseContext.
func (s *ExampleHTTPServer) start(ctx context.Context) (err error) {
	s.httpServer.BaseContext = func(_ net.Listener) context.Context {
		return ctx
	}

	err = s.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

// backup runs an imaginary backup routine
func (s *ExampleHTTPServer) backup() (err error) {
	log.Printf("ExampleHTTPServer.backup: backing up some imaginary stuff")
	time.Sleep(time.Second * 2)
	log.Printf("ExampleHTTPServer.backup: backuped everything")
	return nil
}

// shutdown shuts down the http server so no new connectionos
// are accepted and running connections will have time to finish
func (s *ExampleHTTPServer) shutdown() (err error) {
	log.Printf("ExampleHTTPServer.shutdown: shutting down http server")
	return s.httpServer.Shutdown(context.Background())
}

// close runs shutdown and backup and errors.Join their errors and returns
func (s *ExampleHTTPServer) close(ctx context.Context) (err error) {
	<-ctx.Done()
	err = errors.Join(s.shutdown())
	err = errors.Join(s.backup())
	return
}
