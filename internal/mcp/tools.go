package mcp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Tools handles MCP tool operations
type Tools struct {
	store    *MessageStore
	shutdown <-chan struct{}
}

func NewTools(store *MessageStore, shutdown <-chan struct{}) *Tools {
	return &Tools{
		store:    store,
		shutdown: shutdown,
	}
}

func (t *Tools) Send(args map[string]interface{}) ToolResult {
	from, _ := args["from"].(string)
	to, _ := args["to"].(string)
	content, _ := args["content"].(string)

	if from == "" || to == "" || content == "" {
		return t.errorResult("from, to, and content are required")
	}

	msgID := fmt.Sprintf("msg-%s", uuid.New().String())
	msg := &Message{
		ID:        msgID,
		From:      from,
		To:        to,
		Content:   content,
		Timestamp: time.Now().UnixNano(),
	}

	t.store.StoreMessage(msg)
	ch := t.store.GetOrCreateChannel(to)

	select {
	case ch <- msg:
		// OK
	default:
		t.store.DeleteMessage(msgID)
		return t.errorResult("message queue full for target agent")
	}

	fmt.Printf("[MCP] Message from %s to %s: %s\n", from, to, msgID)

	output := MessageOutput{
		Success:   true,
		MessageID: msgID,
	}
	return t.successResult(output)
}

func (t *Tools) Get(args map[string]interface{}) ToolResult {
	agentName, _ := args["agent_name"].(string)

	if agentName == "" {
		return t.errorResult("agent_name is required")
	}

	ch := t.store.GetOrCreateChannel(agentName)

	select {
	case msg := <-ch:
		fmt.Printf("[MCP] Agent %s received message: %s\n", agentName, msg.ID)
		output := MessageOutput{
			Success:   true,
			MessageID: msg.ID,
			Message:   msg,
		}
		return t.successResult(output)
	case <-t.shutdown:
		return t.errorResult("server shutting down")
	}
}

func (t *Tools) Ack(args map[string]interface{}) ToolResult {
	messageID, _ := args["message_id"].(string)

	if messageID == "" {
		return t.errorResult("message_id is required")
	}

	if !t.store.MessageExists(messageID) {
		return t.errorResult("message not found")
	}

	t.store.DeleteMessage(messageID)
	fmt.Printf("[MCP] Message acknowledged and removed: %s\n", messageID)

	output := MessageOutput{
		Success:   true,
		MessageID: messageID,
	}
	return t.successResult(output)
}

func (t *Tools) successResult(output interface{}) ToolResult {
	data, _ := json.Marshal(output)
	return ToolResult{
		Content: []ContentItem{
			{Type: "text", Text: string(data)},
		},
	}
}

func (t *Tools) errorResult(msg string) ToolResult {
	output := MessageOutput{
		Success: false,
		Error:   msg,
	}
	data, _ := json.Marshal(output)
	return ToolResult{
		Content: []ContentItem{
			{Type: "text", Text: string(data)},
		},
		IsError: true,
	}
}
