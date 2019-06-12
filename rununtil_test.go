package rununtil_test

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/mec07/rununtil"
)

func helperSendSignal(t *testing.T, p *os.Process, sent *bool, signal os.Signal, delay time.Duration) {
	time.Sleep(delay)
	if err := p.Signal(signal); err != nil {
		t.Errorf("unexpected error occurred: %v", err)
	}
	*sent = true
}

func helperMakeFakeRunner(hasBeenShutdown *bool) rununtil.RunnerFunc {
	return rununtil.RunnerFunc(func() rununtil.ShutdownFunc {
		return rununtil.ShutdownFunc(func() {
			*hasBeenShutdown = true
		})
	})
}

func helperMakeMain(hasBeenKilled *bool) func() {
	return func() {
		rununtil.KillSignal(helperMakeFakeRunner(hasBeenKilled))
	}
}

func TestRunUntilKillSignal(t *testing.T) {
	table := []struct {
		name   string
		signal os.Signal
	}{
		{
			name:   "Shutdown from SIGINT",
			signal: syscall.SIGINT,
		},
		{
			name:   "Shutdown from SIGTERM",
			signal: syscall.SIGTERM,
		},
	}
	for _, test := range table {
		t.Run(test.name, func(t *testing.T) {
			var sentSignal bool
			var hasBeenShutdown bool
			p, err := os.FindProcess(os.Getpid())
			if err != nil {
				t.Fatalf("Unexpected error when finding process: %v", err)
			}

			go helperSendSignal(t, p, &sentSignal, test.signal, 1*time.Millisecond)
			rununtil.KillSignal(helperMakeFakeRunner(&hasBeenShutdown))
			if !sentSignal {
				t.Fatal("expected signal to have been sent")
			}
			if !hasBeenShutdown {
				t.Fatal("expected the shutdown function to have been called")
			}
		})
	}
}

func TestRunUntilKilled(t *testing.T) {
	var hasBeenKilled bool
	kill := rununtil.Killed(helperMakeMain(&hasBeenKilled))
	kill()

	// yield control back to scheduler so that killing can actually happen
	time.Sleep(time.Millisecond)
	if !hasBeenKilled {
		t.Fatal("expected main to have been killed")
	}
}
