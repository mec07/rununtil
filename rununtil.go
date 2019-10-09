/*Package rununtil has been created to run a provided function until it has been signalled to stop.

Usage

The main usage of rununtil is to run your main app indefinitely until a SIGINT or SIGTERM signal has been received.
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
		rununtil.AwaitKillSignal(Runner)
	}

The `AwaitKillSignal` function blocks until either a kill signal has been received or `SimulateKillSignal` has been triggered.
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
		rununtil.AwaitKillSignal(NewRunner(logger))
	}

It is of course possible to specify which signals you would like to use to kill your application using the `AwaitKillSignals` function, for example:
	rununtil.AwaitKillSignals([]os.Signal{syscall.SIGKILL, syscall.SIGHUP, syscall.SIGINT}, NewRunner(logger))

For testing purposes you may want to run your main function, which is using `rununtil.AwaitKillSignal`, and then kill it by simulating sending a kill signal when you're done with your tests. To aid with this you can:
	go main()
	... do your tests ...
	rununtil.SimulateKillSignal()

The `SimulateKillSignal` function kills all of the `RunnerFuncs` that you may have started in your main function.

The old functions `KillSignal`, `Signals` and `Killed` are still here (for backwards compatibility), but they have been deprecated.
Please use `AwaitKillSignal` instead of `KillSignal`, `AwaitKillSignals` instead of `Signals`, and `SimulateKillSignal` instead of `Killed` (now you can just rain main and then issue the simulated kill signal).
*/
package rununtil

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type canceller struct {
	signals map[string]chan struct{}
	mux     sync.Mutex
}

func (canc *canceller) addChannel(key string, c chan struct{}) {
	canc.mux.Lock()
	canc.signals[key] = c
	canc.mux.Unlock()
}

func (canc *canceller) removeChannel(key string) {
	canc.mux.Lock()
	delete(canc.signals, key)
	canc.mux.Unlock()
}

func (canc *canceller) cancelAll() {
	canc.mux.Lock()
	for key := range canc.signals {
		close(canc.signals[key])
	}
	canc.mux.Unlock()
}

var globalCanceller canceller

func init() {
	globalCanceller.mux.Lock()
	globalCanceller.signals = make(map[string]chan struct{})
	globalCanceller.mux.Unlock()
}

// ShutdownFunc is a function that should be returned by a RunnerFunc which
// gracefully shuts down whatever is being run.
type ShutdownFunc func()

// RunnerFunc is a function that sets off the worker go routines and returns
// a function which can shutdown those worker go routines.
type RunnerFunc func() ShutdownFunc

// AwaitKillSignal runs the provided runner functions until it receives a kill
// signal, SIGINT or SIGTERM, at which point it executes the graceful shutdown
// functions.
func AwaitKillSignal(runnerFuncs ...RunnerFunc) {
	AwaitKillSignals([]os.Signal{syscall.SIGINT, syscall.SIGTERM}, runnerFuncs...)
}

// AwaitKillSignals runs the provided runner function until the specified
// signals have been recieved, at which point it executes the graceful shutdown
// functions.
func AwaitKillSignals(signals []os.Signal, runnerFuncs ...RunnerFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, signals...)

	finish := make(chan struct{})
	uuid := uuid.New()
	globalCanceller.addChannel(uuid.String(), finish)

	for _, runner := range runnerFuncs {
		shutdown := runner()
		defer shutdown()
	}

	// Wait for a kill signal
	select {
	case <-c:
		break
	case <-finish:
		break
	}
	globalCanceller.removeChannel(uuid.String())
}

// SimulateKillSignal will stop all the awaits in the same way that a kill
// signal would stop them.
func SimulateKillSignal() {
	globalCanceller.cancelAll()
}

// KillSignal runs the provided runner function until it receives a kill signal,
// SIGINT or SIGTERM, at which point it executes the graceful shutdown function.
// Deprecated. Please use AwaitKillSignal.
func KillSignal(runner RunnerFunc) {
	AwaitKillSignal(runner)
}

// Signals runs the provided runner function until the specified signals have
// been recieved.
// Deprecated. Please use AwaitKillSignals.
func Signals(runner RunnerFunc, signals ...os.Signal) {
	AwaitKillSignals(signals, runner)
}

// Killed is used for testing a function that is using rununtil.KillSignal.
// It runs the function provided and sends a SIGINT signal to kill it when
// the returned context.CancelFunc is executed. A sample usage of this could be:
//	kill := rununtil.Killed(main)
//	... do some stuff, e.g. send some requests to the webserver ...
//	kill()
//
// where main is a function that is using rununtil.KillSignal.
// Deprecated. Please just run your main function and use
// rununtil.SimulateKillSignal.
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

	SimulateKillSignal()
}
