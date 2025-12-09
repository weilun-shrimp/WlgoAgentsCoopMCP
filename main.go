package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/weilun-shrimp/wlgo_svc_lifecycle_mgr"
	"wlgoagentscoopmcp/internal/kernel"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		panic(fmt.Sprintf("%s-%s", "Load .env file error", err.Error()))
	}

	manager := wlgo_svc_lifecycle_mgr.NewManager()

	defer func() {
		result := manager.Rollback()
		if result.GetError() != nil {
			log.Printf("Rollback failed: %v", result.GetError())
		}
	}()

	httpServer := kernel.NewHttpServer()
	manager.AddService(httpServer)

	// MCP Server for agent communication
	mcpServer := kernel.NewMCPServer()
	manager.AddService(mcpServer)

	// QuitSignal as final service - blocks until SIGINT/SIGTERM received
	quitSignal := kernel.NewQuitSignal()
	manager.AddService(quitSignal)

	result := manager.Start()
	if result.GetError() != nil {
		log.Printf("Startup failed: %v", result.GetError())
		return
	}

	log.Println("Shutdown complete")
}
