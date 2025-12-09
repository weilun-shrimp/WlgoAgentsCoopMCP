package kernel

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/weilun-shrimp/wlgo_svc_lifecycle_mgr"
)

type QuitSignal struct {
	wlgo_svc_lifecycle_mgr.ServiceProvider
	quit chan os.Signal
}

func NewQuitSignal() *QuitSignal {
	return &QuitSignal{}
}

func (s *QuitSignal) GetName() string {
	return "QuitSignal"
}

func (s *QuitSignal) Begin() error {
	s.quit = make(chan os.Signal, 1)
	signal.Notify(s.quit, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("QuitSignal service started, waiting for shutdown signal...")

	// Block until we receive a signal
	sig := <-s.quit

	fmt.Printf("Received signal: %v, initiating shutdown...\n", sig)
	return nil
}

func (s *QuitSignal) End() error {
	if s.quit != nil {
		signal.Stop(s.quit)
		close(s.quit)
	}
	fmt.Println("QuitSignal service stopped")
	return nil
}
