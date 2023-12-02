# Graceful

Package graceful let's you run your programs gracefully managing signal handling, cleanup timeouts and force quit for you.

```go
package main

import (
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

	grace.GoWithContext(runApplication)
	grace.FatalWait()
}
```

For more information on how to use it refer to [examples](./examples/README.md) readme.