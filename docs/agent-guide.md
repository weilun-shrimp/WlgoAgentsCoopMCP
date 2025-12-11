# Agent Communication Guide

This guide provides detailed reference for the MCP tools and how to use them.

## Prerequisites

1. WlgoAgentsCoopMCP server running on `http://localhost:3001`
2. An AI agent that supports MCP (e.g., Claude Code CLI)

## MCP Server Registration

### Claude Code CLI

Add to `.mcp.json` (project root) or `~/.claude.json` (user-level):

```json
{
  "mcpServers": {
    "agents-coop": {
      "type": "http",
      "url": "http://localhost:3001"
    }
  }
}
```

### Other MCP-Compatible Tools

- **URL**: `http://localhost:3001`
- **Transport**: Streamable HTTP (MCP 2025-03-26 spec)

## Tool Reference

### `send`

Send a message to another agent. The message is stored until the target agent retrieves it.

**Parameters:**

| Parameter | Type   | Required | Description                 |
|-----------|--------|----------|-----------------------------|
| `from`    | string | Yes      | Your agent name             |
| `to`      | string | Yes      | Target agent name           |
| `content` | string | Yes      | Message content             |

**Request:**
```json
{
  "from": "developer",
  "to": "reviewer",
  "content": "Implemented login feature. Please review src/auth.go"
}
```

**Response:**
```json
{
  "success": true,
  "message_id": "msg-550e8400-e29b-41d4-a716-446655440000"
}
```

---

### `get`

Wait for and receive a message. **This call blocks until a message arrives.**

**Parameters:**

| Parameter    | Type   | Required | Description       |
|--------------|--------|----------|-------------------|
| `agent_name` | string | Yes      | Your agent name   |

**Request:**
```json
{
  "agent_name": "reviewer"
}
```

**Response:**
```json
{
  "success": true,
  "message_id": "msg-550e8400-e29b-41d4-a716-446655440000",
  "message": {
    "id": "msg-550e8400-e29b-41d4-a716-446655440000",
    "from": "developer",
    "to": "reviewer",
    "content": "Implemented login feature. Please review src/auth.go",
    "timestamp": 1733745296000000000
  }
}
```

---

### `ack`

Acknowledge a message and remove it from storage. Call this after you've processed a message.

**Parameters:**

| Parameter    | Type   | Required | Description            |
|--------------|--------|----------|------------------------|
| `message_id` | string | Yes      | The message ID to ack  |

**Request:**
```json
{
  "message_id": "msg-550e8400-e29b-41d4-a716-446655440000"
}
```

**Response:**
```json
{
  "success": true,
  "message_id": "msg-550e8400-e29b-41d4-a716-446655440000"
}
```

## Error Reference

| Error | Cause |
|-------|-------|
| `"from, to, and content are required"` | Missing parameters in `send` |
| `"agent_name is required"` | Missing parameter in `get` |
| `"message_id is required"` | Missing parameter in `ack` |
| `"message not found"` | Invalid message ID in `ack` |
| `"message queue full for target agent"` | Queue overflow (max 100 messages) |
| `"context cancelled while waiting for message"` | Request timeout |

## Agent Identity

Each agent needs a unique name. Define this in the agent's system prompt:

```
You are "developer" in a multi-agent system.

When sending: use "developer" as your `from` field
When receiving: use "developer" as your `agent_name`
```

**Naming conventions:**
- Use lowercase with hyphens: `developer`, `code-reviewer`, `test-runner`
- Keep names unique within your pipeline
- Document all agent names so they can find each other
