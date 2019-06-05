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

func helperFakeRunner() func() {
	return func() {}
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
			p, err := os.FindProcess(os.Getpid())
			if err != nil {
				t.Fatalf("Unexpected error when finding process: %v", err)
			}

			go helperSendSignal(t, p, &sentSignal, test.signal, 1*time.Millisecond)
			rununtil.KillSignal(helperFakeRunner)
			if !sentSignal {
				t.Fatal("expected signal to have been sent")
			}
		})
	}
}
