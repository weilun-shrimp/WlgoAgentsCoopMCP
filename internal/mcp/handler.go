package mcp

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// Handler handles MCP HTTP requests
type Handler struct {
	store    *MessageStore
	shutdown <-chan struct{}
}

func NewHandler(store *MessageStore, shutdown <-chan struct{}) *Handler {
	return &Handler{
		store:    store,
		shutdown: shutdown,
	}
}

// HandleHTTP handles MCP JSON-RPC requests over HTTP POST
func (h *Handler) HandleHTTP(c *fiber.Ctx) error {
	var req JSONRPCRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(400).JSON(JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      nil,
			Error: &RPCError{
				Code:    -32700,
				Message: "Parse error",
			},
		})
	}

	fmt.Printf("[MCP] %s (id=%v)\n", req.Method, req.ID)

	response := h.handleRequest(&req)
	if response == nil {
		// For notifications, return 202 Accepted with no body
		return c.SendStatus(202)
	}

	return c.JSON(response)
}

func (h *Handler) handleRequest(req *JSONRPCRequest) *JSONRPCResponse {
	switch req.Method {
	case "initialize":
		return h.handleInitialize(req)
	case "notifications/initialized":
		return nil
	case "tools/list":
		return h.handleToolsList(req)
	case "tools/call":
		return h.handleToolCall(req)
	default:
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &RPCError{
				Code:    -32601,
				Message: "Method not found",
			},
		}
	}
}

func (h *Handler) handleInitialize(req *JSONRPCRequest) *JSONRPCResponse {
	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{},
		},
		ServerInfo: ServerInfo{
			Name:    "WlgoAgentsCoopMCP",
			Version: "v1.0.0",
		},
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

func (h *Handler) handleToolsList(req *JSONRPCRequest) *JSONRPCResponse {
	tools := []Tool{
		{
			Name:        "send",
			Description: "Send a message to another agent. The message will be queued until the target agent retrieves it.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"from":    {Type: "string", Description: "The sender agent name"},
					"to":      {Type: "string", Description: "The target agent name to receive the message"},
					"content": {Type: "string", Description: "The message content"},
				},
				Required: []string{"from", "to", "content"},
			},
		},
		{
			Name:        "get",
			Description: "Wait for and retrieve a message sent to this agent. Returns null message if timeout is reached.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"agent_name": {Type: "string", Description: "The agent name waiting for messages"},
					"timeout":    {Type: "number", Description: "Timeout in seconds to wait for a message (default: 30)"},
				},
				Required: []string{"agent_name"},
			},
		},
		{
			Name:        "ack",
			Description: "Acknowledge a message and remove it from storage.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"message_id": {Type: "string", Description: "The message ID to acknowledge and remove"},
				},
				Required: []string{"message_id"},
			},
		},
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  ToolsListResult{Tools: tools},
	}
}

func (h *Handler) handleToolCall(req *JSONRPCRequest) *JSONRPCResponse {
	var params ToolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &RPCError{
				Code:    -32602,
				Message: "Invalid params",
			},
		}
	}

	tools := NewTools(h.store, h.shutdown)
	var result ToolResult

	switch params.Name {
	case "send":
		result = tools.Send(params.Arguments)
	case "get":
		result = tools.Get(params.Arguments)
	case "ack":
		result = tools.Ack(params.Arguments)
	default:
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &RPCError{
				Code:    -32602,
				Message: fmt.Sprintf("Unknown tool: %s", params.Name),
			},
		}
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}
