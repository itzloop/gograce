# Examples

## Blog examples

These are the examples used in the blog post

- [Basic Example](./blog/01_done_channel_wait_group/main.go) that i used in the blog post as a base and improved it step
  by step.
- [Example with Done Channel and WaitGroup with Force Quit](./blog/02_done_channel_wait_group_force/main.go)
- [Example with Timeouts](./blog/03_done_channel_wait_group_force_timeout/main.go)
- [Example with context.Context instead of Done Channel](./blog/04_context_wait_group_force_timeout/main.go)
- [Example with errgroup.Group instead of WaitGroups](./blog/05_context_errgroup_force_timeout/main.go)

## Library Example

- [HTTP Server](lib/http_example/main.go)

Creates a simples http.Server with:
- `/long-running-job` that sleeps for 10 seconds to demonstrate all active connections will finish before termination.
- `/hello` that will write `hi` back :).

### Running the example

```bash
$ go run examples/lib/http_example/main.go
```

You now have a HTTP server that has to endpoints:
1. `/hello` which simply writes `hi` back.
2. `log-runing-job` which is simple `time.Sleep` and then writes `done` back.

Now when you send a HTTP request to `/log-running-job`, and press `ctrl+c` in server, the code will wait for all requests to finish and then terminates.
