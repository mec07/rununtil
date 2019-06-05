/*Package rununtil has been created to run a provided function until it has been signalled to stop.

Usage

The main usage of rununtil is to run a webserver or other function until a kill signal has been received.
The runner function can do some setup but it should not run indefinitely, instead it should start go routines which can run in the background.
The runner function should return a graceful shutdown function that will be called once the signal has been received.
For example:
	func Runner() func() {
		r := chi.NewRouter()
		r.Get("/healthz", healthzHandler)
		httpServer := &http.Server{Addr: "localhost:8080", Handler: r}
		go runHTTPServer(httpServer)

		return func() {
			if err := httpServer.Shutdown(context.Background()); err != nil {
				log.Error().Err(err).Msg("error occurred while shutting down http server")
			}
		}
	}
	func runHTTPServer(srv *http.Server) {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Stack().Err(err).Msg("ListenAndServe")
		}
	}

	func main() {
		rununtil.KillSignal(Runner)
	}

A nice pattern is to create a function that takes in the various depencies required, e.g. configuration, logger, private key, etc., and returns a runner function, e.g.
	func NewRunner(log *zerolog.Logger) func() func() {
		return Runner() func() {
			...all that was in the Runner function above...
		}
	}

	func main() {
		logger, err := setupLogger()
		if err != nil {
			return
		}
		rununtil.KillSignal(NewRunner(logger))
	}
*/
package rununtil

import (
	"os"
	"os/signal"
	"syscall"
)

// KillSignal runs the provided runner function until it receives a kill signal,
// at which point it executes the graceful shutdown function.
func KillSignal(runner func() func()) {
	Signals(runner, syscall.SIGINT, syscall.SIGTERM)
}

// Signals runs the provided runner function until the specified signals have
// been recieved.
func Signals(runner func() func(), signals ...os.Signal) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, signals...)

	gracefulShutdown := runner()
	defer gracefulShutdown()

	// Wait for a kill signal
	<-c
}
