package kernel

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/weilun-shrimp/wlgo_svc_lifecycle_mgr"

	"wlgoagentscoopmcp/internal/mcp"
)

type HttpServer struct {
	wlgo_svc_lifecycle_mgr.ServiceProvider
	app        *fiber.App
	mcpHandler *mcp.Handler
	shutdown   chan struct{}
}

func NewHttpServer() *HttpServer {
	shutdown := make(chan struct{})
	store := mcp.NewMessageStore()

	return &HttpServer{
		mcpHandler: mcp.NewHandler(store, shutdown),
		shutdown:   shutdown,
	}
}

func (s *HttpServer) GetName() string {
	return "HttpServer"
}

func (s *HttpServer) Begin() error {
	s.app = fiber.New(fiber.Config{
		AppName: os.Getenv("APP_NAME"),
	})

	s.registerRoutes()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	go func() {
		if err := s.app.Listen(":" + port); err != nil {
			fmt.Printf("HttpServer error: %v\n", err)
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		}
	}()

	fmt.Printf("HttpServer started on port %s\n", port)
	return nil
}

func (s *HttpServer) End() error {
	if s.app != nil {
		fmt.Println("HttpServer shutting down...")
		close(s.shutdown)
		return s.app.ShutdownWithTimeout(10 * time.Second)
	}
	return nil
}

func (s *HttpServer) registerRoutes() {
	s.app.Get("/ping", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"msg": "Pong",
		})
	})

	// WebSocket MCP endpoint
	s.app.Use("/mcp", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	s.app.Get("/mcp", websocket.New(s.mcpHandler.HandleWebSocket))
}
