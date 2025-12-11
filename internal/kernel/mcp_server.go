package kernel

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/weilun-shrimp/wlgo_svc_lifecycle_mgr"
)

// Message represents a message between agents
type Message struct {
	ID        string `json:"id"`
	From      string `json:"from"`
	To        string `json:"to"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

// MessageStore manages messages between agents
type MessageStore struct {
	mu       sync.RWMutex
	messages map[string]*Message      // key: message ID (for ack/removal)
	channels map[string]chan *Message // key: agent name (receiver)
}

func NewMessageStore() *MessageStore {
	return &MessageStore{
		messages: make(map[string]*Message),
		channels: make(map[string]chan *Message),
	}
}

// MCPServer implements ServiceProvider for the MCP server
type MCPServer struct {
	wlgo_svc_lifecycle_mgr.ServiceProvider
	server     *mcp.Server
	httpServer *http.Server
	msgStore   *MessageStore
}

func NewMCPServer() *MCPServer {
	return &MCPServer{
		msgStore: NewMessageStore(),
	}
}

func (s *MCPServer) GetName() string {
	return "MCPServer"
}

func (s *MCPServer) Begin() error {
	// Create MCP server
	s.server = mcp.NewServer(&mcp.Implementation{
		Name:    "WlgoAgentsCoopMCP",
		Version: "v1.0.0",
	}, nil)

	// Register tools
	s.registerTools()

	port := os.Getenv("MCP_PORT")
	if port == "" {
		port = "3001"
	}

	// Create HTTP handler using Streamable HTTP transport (replaces deprecated SSE)
	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return s.server
	}, &mcp.StreamableHTTPOptions{
		Stateless: true, // Simpler mode - no session ID validation needed
	})

	s.httpServer = &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("MCPServer error: %v\n", err)
			// Send SIGINT to current process to trigger graceful shutdown
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		}
	}()

	fmt.Printf("MCPServer started on port %s\n", port)
	return nil
}

func (s *MCPServer) End() error {
	if s.httpServer != nil {
		fmt.Println("MCPServer shutting down...")
		return s.httpServer.Shutdown(context.Background())
	}
	return nil
}

// Tool input/output types

// SendInput is the input for the send tool
type SendInput struct {
	From    string `json:"from" jsonschema_description:"The sender agent name"`
	To      string `json:"to" jsonschema_description:"The target agent name to receive the message"`
	Content string `json:"content" jsonschema_description:"The message content"`
}

// GetInput is the input for the get tool
type GetInput struct {
	AgentName string `json:"agent_name" jsonschema_description:"The agent name waiting for messages"`
}

// AckInput is the input for the ack tool
type AckInput struct {
	MessageID string `json:"message_id" jsonschema_description:"The message ID to acknowledge and remove"`
}

// MessageOutput is the common output for message operations
type MessageOutput struct {
	Success   bool     `json:"success"`
	MessageID string   `json:"message_id,omitempty"`
	Message   *Message `json:"message,omitempty"`
	Error     string   `json:"error,omitempty"`
}

func (s *MCPServer) registerTools() {
	// Tool: send - Send a message to another agent
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "send",
		Description: "Send a message to another agent. The message will be queued until the target agent retrieves it.",
	}, s.send)

	// Tool: get - Get messages sent to this agent
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get",
		Description: "Wait for and retrieve a message sent to this agent. Blocks until a message is available.",
	}, s.get)

	// Tool: ack - Acknowledge and remove a message
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "ack",
		Description: "Acknowledge a message and remove it from storage.",
	}, s.ack)
}

func (s *MCPServer) send(ctx context.Context, req *mcp.CallToolRequest, input SendInput) (*mcp.CallToolResult, MessageOutput, error) {
	// Input validation
	if input.From == "" || input.To == "" || input.Content == "" {
		return nil, MessageOutput{
			Success: false,
			Error:   "from, to, and content are required",
		}, nil
	}

	s.msgStore.mu.Lock()

	// Create message with UUID
	msgID := fmt.Sprintf("msg-%s", uuid.New().String())
	msg := &Message{
		ID:        msgID,
		From:      input.From,
		To:        input.To,
		Content:   input.Content,
		Timestamp: time.Now().UnixNano(),
	}

	// Store message for later ack/removal
	s.msgStore.messages[msgID] = msg

	// Get or create channel for the target agent
	ch, exists := s.msgStore.channels[input.To]
	if !exists {
		ch = make(chan *Message, 100)
		s.msgStore.channels[input.To] = ch
	}
	s.msgStore.mu.Unlock()

	// Send message to channel with error handling for full channel
	select {
	case ch <- msg:
		// OK
	default:
		// Remove from store if channel is full
		s.msgStore.mu.Lock()
		delete(s.msgStore.messages, msgID)
		s.msgStore.mu.Unlock()
		return nil, MessageOutput{
			Success: false,
			Error:   "message queue full for target agent",
		}, nil
	}

	fmt.Printf("[MCP] Message from %s to %s: %s\n", input.From, input.To, msgID)

	return nil, MessageOutput{
		Success:   true,
		MessageID: msgID,
	}, nil
}

func (s *MCPServer) get(ctx context.Context, req *mcp.CallToolRequest, input GetInput) (*mcp.CallToolResult, MessageOutput, error) {
	// Input validation
	if input.AgentName == "" {
		return nil, MessageOutput{
			Success: false,
			Error:   "agent_name is required",
		}, nil
	}

	s.msgStore.mu.Lock()
	ch, exists := s.msgStore.channels[input.AgentName]
	if !exists {
		ch = make(chan *Message, 100)
		s.msgStore.channels[input.AgentName] = ch
	}
	s.msgStore.mu.Unlock()

	// Wait for message
	select {
	case msg := <-ch:
		fmt.Printf("[MCP] Agent %s received message: %s\n", input.AgentName, msg.ID)
		return nil, MessageOutput{
			Success:   true,
			MessageID: msg.ID,
			Message:   msg,
		}, nil
	case <-ctx.Done():
		return nil, MessageOutput{
			Success: false,
			Error:   "context cancelled while waiting for message",
		}, nil
	}
}

func (s *MCPServer) ack(ctx context.Context, req *mcp.CallToolRequest, input AckInput) (*mcp.CallToolResult, MessageOutput, error) {
	// Input validation
	if input.MessageID == "" {
		return nil, MessageOutput{
			Success: false,
			Error:   "message_id is required",
		}, nil
	}

	s.msgStore.mu.Lock()
	defer s.msgStore.mu.Unlock()

	_, exists := s.msgStore.messages[input.MessageID]
	if !exists {
		return nil, MessageOutput{
			Success: false,
			Error:   "message not found",
		}, nil
	}

	// Remove message from store
	delete(s.msgStore.messages, input.MessageID)
	fmt.Printf("[MCP] Message acknowledged and removed: %s\n", input.MessageID)

	return nil, MessageOutput{
		Success:   true,
		MessageID: input.MessageID,
	}, nil
}
