# Gograceful
> Gograceful let's you run your programs gracefully managing signal handling, cleanup timeouts and force quit for you.

![Red Sus](./.github/sus.png)

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/itzloop/gograceful)](https://goreportcard.com/report/github.com/itzloop/gograceful)
[![PkgGoDev](https://pkg.go.dev/badge/mod/github.com/itzloop/gograceful)](https://pkg.go.dev/mod/github.com/itzloop/gograceful)

## Usage
```go
package main

import (
	"context"
	"github.com/itzloop/graceful"
	"time"
)

func main() {
	grace := graceful.NewGraceful(graceful.Options{
		Timeout:       10 * time.Second,
		NoForceQuit:   false,
		MaxGoRoutines: 0,
		Signals:       nil,
	})

	app := App{}
	grace.GoWithContext(app.Start)
	grace.GoWithContext(app.Close)
	grace.FatalWait()
}

type App struct{}

func (app *App) Start(ctx context.Context) error {
	// run stuff
}


func (app *App) Close(ctx context.Context) error {
	<-ctx.Done()
	// do cleanup
}
```

For more information on how to use it refer to [examples](./examples/README.md) readme.

## Testing

TODO