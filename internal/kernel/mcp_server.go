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
	Type      string `json:"type"` // "commit" or "feedback"
	Received  bool   `json:"received"`
	Timestamp int64  `json:"timestamp"`
}

// MessageStore manages messages between agents
type MessageStore struct {
	mu       sync.RWMutex
	messages map[string]*Message // key: message ID
	// Channels for agents to wait for messages
	commitChannels   map[string]chan *Message // key: agent name (receiver)
	feedbackChannels map[string]chan *Message // key: agent name (receiver)
}

func NewMessageStore() *MessageStore {
	return &MessageStore{
		messages:         make(map[string]*Message),
		commitChannels:   make(map[string]chan *Message),
		feedbackChannels: make(map[string]chan *Message),
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

	// Create HTTP handler using SSE transport
	handler := mcp.NewSSEHandler(func(r *http.Request) *mcp.Server {
		return s.server
	}, nil)

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

// CommitToNextInput is the input for commit_to_next tool
type CommitToNextInput struct {
	From    string `json:"from" jsonschema_description:"The agent name sending the commit"`
	To      string `json:"to" jsonschema_description:"The target agent name to receive the commit"`
	Content string `json:"content" jsonschema_description:"The commit message content"`
}

// FeedbackToPreviousInput is the input for feedback_to_previous tool
type FeedbackToPreviousInput struct {
	From    string `json:"from" jsonschema_description:"The agent name sending the feedback"`
	To      string `json:"to" jsonschema_description:"The target agent name to receive the feedback"`
	Content string `json:"content" jsonschema_description:"The feedback message content"`
}

// GetPreviousCommitInput is the input for get_previous_commit tool
type GetPreviousCommitInput struct {
	AgentName string `json:"agent_name" jsonschema_description:"The agent name waiting for a commit"`
}

// GetNextFeedbackInput is the input for get_next_feedback tool
type GetNextFeedbackInput struct {
	AgentName string `json:"agent_name" jsonschema_description:"The agent name waiting for feedback"`
}

// AckMessageInput is the input for ack_message tool
type AckMessageInput struct {
	MessageID string `json:"message_id" jsonschema_description:"The message ID to acknowledge as received"`
}

// MessageOutput is the common output for message operations
type MessageOutput struct {
	Success   bool     `json:"success"`
	MessageID string   `json:"message_id,omitempty"`
	Message   *Message `json:"message,omitempty"`
	Error     string   `json:"error,omitempty"`
}

func (s *MCPServer) registerTools() {
	// Tool: commit_to_next - Send a commit message to the next agent
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "commit_to_next",
		Description: "Send a commit message to the next agent in the pipeline. The message will be stored until the target agent retrieves it.",
	}, s.commitToNext)

	// Tool: feedback_to_previous - Send feedback to the previous agent
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "feedback_to_previous",
		Description: "Send feedback/rollback message to the previous agent in the pipeline. The message will be stored until the target agent retrieves it.",
	}, s.feedbackToPrevious)

	// Tool: get_previous_commit - Get commit message from previous agent
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get_previous_commit",
		Description: "Wait for and retrieve a commit message from the previous agent. Blocks until a message is available.",
	}, s.getPreviousCommit)

	// Tool: get_next_feedback - Get feedback from next agent
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get_next_feedback",
		Description: "Wait for and retrieve a feedback message from the next agent. Blocks until a message is available.",
	}, s.getNextFeedback)

	// Tool: ack_message - Acknowledge message as received
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "ack_message",
		Description: "Acknowledge that a message has been received and read. This notifies the sender that their message was delivered.",
	}, s.ackMessage)
}

func (s *MCPServer) commitToNext(ctx context.Context, req *mcp.CallToolRequest, input CommitToNextInput) (*mcp.CallToolResult, MessageOutput, error) {
	// Input validation
	if input.From == "" || input.To == "" || input.Content == "" {
		return nil, MessageOutput{
			Success: false,
			Error:   "from, to, and content are required",
		}, nil
	}

	s.msgStore.mu.Lock()

	// Create message with UUID
	msgID := fmt.Sprintf("commit-%s", uuid.New().String())
	msg := &Message{
		ID:        msgID,
		From:      input.From,
		To:        input.To,
		Content:   input.Content,
		Type:      "commit",
		Received:  false,
		Timestamp: time.Now().UnixNano(),
	}
	s.msgStore.messages[msgID] = msg

	// Get or create channel for the target agent
	ch, exists := s.msgStore.commitChannels[input.To]
	if !exists {
		ch = make(chan *Message, 100)
		s.msgStore.commitChannels[input.To] = ch
	}
	s.msgStore.mu.Unlock()

	// Send message to channel with error handling for full channel
	select {
	case ch <- msg:
		// OK
	default:
		return nil, MessageOutput{
			Success: false,
			Error:   "message queue full for target agent",
		}, nil
	}

	fmt.Printf("[MCP] Commit from %s to %s: %s\n", input.From, input.To, msgID)

	return nil, MessageOutput{
		Success:   true,
		MessageID: msgID,
	}, nil
}

func (s *MCPServer) feedbackToPrevious(ctx context.Context, req *mcp.CallToolRequest, input FeedbackToPreviousInput) (*mcp.CallToolResult, MessageOutput, error) {
	// Input validation
	if input.From == "" || input.To == "" || input.Content == "" {
		return nil, MessageOutput{
			Success: false,
			Error:   "from, to, and content are required",
		}, nil
	}

	s.msgStore.mu.Lock()

	// Create message with UUID
	msgID := fmt.Sprintf("feedback-%s", uuid.New().String())
	msg := &Message{
		ID:        msgID,
		From:      input.From,
		To:        input.To,
		Content:   input.Content,
		Type:      "feedback",
		Received:  false,
		Timestamp: time.Now().UnixNano(),
	}
	s.msgStore.messages[msgID] = msg

	// Get or create channel for the target agent
	ch, exists := s.msgStore.feedbackChannels[input.To]
	if !exists {
		ch = make(chan *Message, 100)
		s.msgStore.feedbackChannels[input.To] = ch
	}
	s.msgStore.mu.Unlock()

	// Send message to channel with error handling for full channel
	select {
	case ch <- msg:
		// OK
	default:
		return nil, MessageOutput{
			Success: false,
			Error:   "message queue full for target agent",
		}, nil
	}

	fmt.Printf("[MCP] Feedback from %s to %s: %s\n", input.From, input.To, msgID)

	return nil, MessageOutput{
		Success:   true,
		MessageID: msgID,
	}, nil
}

func (s *MCPServer) getPreviousCommit(ctx context.Context, req *mcp.CallToolRequest, input GetPreviousCommitInput) (*mcp.CallToolResult, MessageOutput, error) {
	// Input validation
	if input.AgentName == "" {
		return nil, MessageOutput{
			Success: false,
			Error:   "agent_name is required",
		}, nil
	}

	s.msgStore.mu.Lock()
	ch, exists := s.msgStore.commitChannels[input.AgentName]
	if !exists {
		ch = make(chan *Message, 100)
		s.msgStore.commitChannels[input.AgentName] = ch
	}
	s.msgStore.mu.Unlock()

	// Wait for message
	select {
	case msg := <-ch:
		fmt.Printf("[MCP] Agent %s received commit: %s\n", input.AgentName, msg.ID)
		return nil, MessageOutput{
			Success:   true,
			MessageID: msg.ID,
			Message:   msg,
		}, nil
	case <-ctx.Done():
		return nil, MessageOutput{
			Success: false,
			Error:   "context cancelled while waiting for commit",
		}, nil
	}
}

func (s *MCPServer) getNextFeedback(ctx context.Context, req *mcp.CallToolRequest, input GetNextFeedbackInput) (*mcp.CallToolResult, MessageOutput, error) {
	// Input validation
	if input.AgentName == "" {
		return nil, MessageOutput{
			Success: false,
			Error:   "agent_name is required",
		}, nil
	}

	s.msgStore.mu.Lock()
	ch, exists := s.msgStore.feedbackChannels[input.AgentName]
	if !exists {
		ch = make(chan *Message, 100)
		s.msgStore.feedbackChannels[input.AgentName] = ch
	}
	s.msgStore.mu.Unlock()

	// Wait for message
	select {
	case msg := <-ch:
		fmt.Printf("[MCP] Agent %s received feedback: %s\n", input.AgentName, msg.ID)
		return nil, MessageOutput{
			Success:   true,
			MessageID: msg.ID,
			Message:   msg,
		}, nil
	case <-ctx.Done():
		return nil, MessageOutput{
			Success: false,
			Error:   "context cancelled while waiting for feedback",
		}, nil
	}
}

func (s *MCPServer) ackMessage(ctx context.Context, req *mcp.CallToolRequest, input AckMessageInput) (*mcp.CallToolResult, MessageOutput, error) {
	// Input validation
	if input.MessageID == "" {
		return nil, MessageOutput{
			Success: false,
			Error:   "message_id is required",
		}, nil
	}

	s.msgStore.mu.Lock()
	defer s.msgStore.mu.Unlock()

	msg, exists := s.msgStore.messages[input.MessageID]
	if !exists {
		return nil, MessageOutput{
			Success: false,
			Error:   "message not found",
		}, nil
	}

	msg.Received = true
	fmt.Printf("[MCP] Message acknowledged: %s\n", input.MessageID)

	return nil, MessageOutput{
		Success:   true,
		MessageID: input.MessageID,
		Message:   msg,
	}, nil
}
