package mcp

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/contrib/websocket"
)

// Handler handles MCP WebSocket connections
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

func (h *Handler) HandleWebSocket(c *websocket.Conn) {
	defer c.Close()

	fmt.Printf("[MCP] WebSocket client connected: %s\n", c.RemoteAddr())

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			fmt.Printf("[MCP] WebSocket read error: %v\n", err)
			return
		}

		var req JSONRPCRequest
		if err := json.Unmarshal(msg, &req); err != nil {
			h.sendError(c, nil, -32700, "Parse error", nil)
			continue
		}

		response := h.handleRequest(&req)
		if response != nil {
			data, _ := json.Marshal(response)
			if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
				fmt.Printf("[MCP] WebSocket write error: %v\n", err)
				return
			}
		}
	}
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
			Description: "Wait for and retrieve a message sent to this agent. Blocks until a message is available.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"agent_name": {Type: "string", Description: "The agent name waiting for messages"},
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

func (h *Handler) sendError(c *websocket.Conn, id interface{}, code int, message string, data interface{}) {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	respData, _ := json.Marshal(response)
	c.WriteMessage(websocket.TextMessage, respData)
}
