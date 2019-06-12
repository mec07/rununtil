/*Package rununtil has been created to run a provided function until it has been signalled to stop.

Usage

The main usage of rununtil is to run a webserver or other function until a SIGINT or SIGTERM signal has been received.
The runner function can do some setup but it should not run indefinitely, instead it should start go routines which can run in the background.
The runner function should return a graceful shutdown function that will be called once the signal has been received.
For example:
	func Runner() rununtil.ShutdownFunc {
		r := chi.NewRouter()
		r.Get("/healthz", healthzHandler)
		httpServer := &http.Server{Addr: ":8080", Handler: r}
		go runHTTPServer(httpServer)

		return rununtil.ShutdownFunc(func() {
			if err := httpServer.Shutdown(context.Background()); err != nil {
				log.Error().Err(err).Msg("error occurred while shutting down http server")
			}
		})
	}

	func runHTTPServer(srv *http.Server) {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Stack().Err(err).Msg("ListenAndServe")
		}
	}

	func main() {
		rununtil.KillSignal(Runner)
	}

A nice pattern is to create a function that takes in the various depencies required, for example, a logger (but could be anything, e.g. configs, database, etc.), and returns a runner function:
	func NewRunner(log *zerolog.Logger) rununtil.RunnerFunc {
		return rununtil.RunnerFunc(func() rununtil.ShutdownFunc {
			r := chi.NewRouter()
			r.Get("/healthz", healthzHandler)
			httpServer := &http.Server{Addr: ":8080", Handler: r}
			go runHTTPServer(httpServer, log)

			return rununtil.ShutdownFunc(func() {
				if err := httpServer.Shutdown(context.Background()); err != nil {
					log.Error().Err(err).Msg("error occurred while shutting down http server")
				}
			})
		})
	}

	func runHTTPServer(srv *http.Server, log *zerolog.Logger) {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Stack().Err(err).Msg("ListenAndServe")
		}
	}

	func main() {
		logger, err := setupLogger()
		if err != nil {
			return
		}
		rununtil.KillSignal(NewRunner(logger))
	}

It is of course possible to specify which signals you would like to use to kill your application using the `Signals` function, for example:
	rununtil.Signals(NewRunner(logger), syscall.SIGKILL, syscall.SIGHUP, syscall.SIGINT)

For testing purposes you may want to run your main function, which is using `rununtil.KillSignal`, and send it a kill signal when you're done with your tests. To aid with this you can use:
	kill := rununtil.Killed(main)

where `kill` is a function that sends a kill signal to the main function when executed (its type is context.CancelFunc).
*/
package rununtil

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
)

// ShutdownFunc is a function that should be returned by a RunnerFunc which
// gracefully shuts down whatever is being run.
type ShutdownFunc func()

// RunnerFunc is a function that sets off the worker go routines and returns
// a function which can shutdown those worker go routines.
type RunnerFunc func() ShutdownFunc

// KillSignal runs the provided runner function until it receives a kill signal,
// SIGINT or SIGTERM, at which point it executes the graceful shutdown function.
func KillSignal(runner RunnerFunc) {
	Signals(runner, syscall.SIGINT, syscall.SIGTERM)
}

// Signals runs the provided runner function until the specified signals have
// been recieved.
func Signals(runner RunnerFunc, signals ...os.Signal) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, signals...)

	shutdown := runner()
	defer shutdown()

	// Wait for a kill signal
	<-c
}

// Killed is used for testing a function that is using rununtil.KillSignal.
// It runs the function provided and sends a SIGINT signal to kill it when
// the returned context.CancelFunc is executed. A sample usage of this could be:
//	kill := rununtil.Killed(main)
//	... do some stuff, e.g. send some requests to the webserver ...
//	kill()
//
// where main is a function that is using rununtil.KillSignal.
func Killed(main func()) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	go runMain(ctx, main)

	return cancel
}

func runMain(ctx context.Context, main func()) {
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		fmt.Printf("ERROR: %+v\n", errors.Wrap(err, "trying to get PID"))
	}
	go killMainWhenDone(ctx, p)
	main()
}

func killMainWhenDone(ctx context.Context, p *os.Process) {
	<-ctx.Done()

	if err := p.Signal(syscall.SIGINT); err != nil {
		fmt.Printf("ERROR: %+v\n", errors.Wrap(err, "trying to kill main"))
	}
}
