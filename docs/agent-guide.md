# AI Agent Communication Guide

This guide explains how to connect AI agents (like Claude Code CLI) to the WlgoAgentsCoopMCP server and enable inter-agent communication.

## Overview

WlgoAgentsCoopMCP provides an MCP (Model Context Protocol) server that allows AI agents to communicate with each other in a pipeline. Agents can:

- **Commit** work to the next agent in the pipeline
- **Receive commits** from the previous agent
- **Send feedback** to the previous agent
- **Receive feedback** from the next agent

## Prerequisites

1. WlgoAgentsCoopMCP server running (default: `http://localhost:3001`)
2. An AI agent that supports MCP (e.g., Claude Code CLI)

## Connecting to the MCP Server

### Claude Code CLI Configuration

Add the MCP server to your Claude Code configuration file (`.mcp.json` in project root or `~/.claude.json` for user-level):

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

### Other MCP-Compatible Agents

For other agents, configure the MCP Streamable HTTP endpoint:

- **URL**: `http://localhost:3001`
- **Transport**: Streamable HTTP (MCP 2025-03-26 spec)

## Available Tools

Once connected, the agent has access to these MCP tools:

### 1. `commit_to_next`

Send your completed work to the next agent in the pipeline.

**Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `from` | string | Your agent name (identifier) |
| `to` | string | Target agent name to receive the commit |
| `content` | string | The commit message/work content |

**Example:**
```json
{
  "from": "developer-agent",
  "to": "reviewer-agent",
  "content": "Implemented user authentication feature. Files changed: auth.go, middleware.go"
}
```

**Response:**
```json
{
  "success": true,
  "message_id": "commit-550e8400-e29b-41d4-a716-446655440000"
}
```

### 2. `get_previous_commit`

Wait for and receive a commit from the previous agent. This call **blocks** until a message is available.

**Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `agent_name` | string | Your agent name (to identify which commits to receive) |

**Example:**
```json
{
  "agent_name": "reviewer-agent"
}
```

**Response:**
```json
{
  "success": true,
  "message_id": "commit-550e8400-e29b-41d4-a716-446655440000",
  "message": {
    "id": "commit-550e8400-e29b-41d4-a716-446655440000",
    "from": "developer-agent",
    "to": "reviewer-agent",
    "content": "Implemented user authentication feature...",
    "type": "commit",
    "received": false,
    "timestamp": 1733745296000000000
  }
}
```

### 3. `feedback_to_previous`

Send feedback or request changes from the previous agent.

**Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `from` | string | Your agent name |
| `to` | string | Target agent name to receive feedback |
| `content` | string | The feedback content |

**Example:**
```json
{
  "from": "reviewer-agent",
  "to": "developer-agent",
  "content": "Please add input validation to the login function"
}
```

### 4. `get_next_feedback`

Wait for and receive feedback from the next agent. This call **blocks** until feedback is available.

**Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `agent_name` | string | Your agent name |

**Example:**
```json
{
  "agent_name": "developer-agent"
}
```

### 5. `ack_message`

Acknowledge that you have received and read a message. This notifies the sender.

**Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `message_id` | string | The message ID to acknowledge |

**Example:**
```json
{
  "message_id": "commit-550e8400-e29b-41d4-a716-446655440000"
}
```

## Communication Patterns

### Pipeline Pattern (Developer -> Reviewer)

```
┌─────────────────┐                    ┌─────────────────┐
│ Developer Agent │                    │ Reviewer Agent  │
└────────┬────────┘                    └────────┬────────┘
         │                                      │
         │  1. commit_to_next(to: reviewer)     │
         │─────────────────────────────────────>│
         │                                      │
         │                    2. get_previous_commit()
         │                                      │
         │  3. ack_message(message_id)          │
         │<─────────────────────────────────────│
         │                                      │
         │  4. feedback_to_previous(to: dev)    │
         │<─────────────────────────────────────│
         │                                      │
    5. get_next_feedback()                      │
         │                                      │
         │  6. ack_message(message_id)          │
         │─────────────────────────────────────>│
         │                                      │
```

### Example: Developer Agent Workflow

```
1. Do your work
2. Call commit_to_next(from: "dev", to: "reviewer", content: "...")
3. Call get_next_feedback(agent_name: "dev")  // Blocks until feedback
4. Call ack_message(message_id: "...")
5. If feedback requires changes, go back to step 1
6. If approved, task complete
```

### Example: Reviewer Agent Workflow

```
1. Call get_previous_commit(agent_name: "reviewer")  // Blocks until commit
2. Call ack_message(message_id: "...")
3. Review the work
4. Call feedback_to_previous(from: "reviewer", to: "dev", content: "...")
5. Go back to step 1
```

## Multi-Agent Pipeline

You can create longer pipelines with multiple agents:

```
Developer -> Reviewer -> QA -> Deployer
```

Each agent:
- Receives from the previous agent using `get_previous_commit`
- Sends to the next agent using `commit_to_next`
- Can request changes using `feedback_to_previous`
- Can receive change requests using `get_next_feedback`

## Agent Identity

Each agent needs to know its own name to use in the `from` field when sending messages and in `agent_name` when receiving. There are several ways to configure this:

### Option 1: System Prompt (Recommended)

Include the agent name in your agent's system prompt:

```
You are "developer-agent" in a multi-agent pipeline.

When communicating with other agents:
- Use "developer-agent" as your `from` field
- Send commits to "reviewer-agent" using commit_to_next
- Listen for feedback using get_next_feedback with agent_name="developer-agent"
```

### Option 2: Environment Variable

Set an environment variable that your agent reads:

```bash
export AGENT_NAME=developer-agent
```

Then reference this in your agent's configuration or prompt.

### Agent Name Conventions

- Use lowercase with hyphens: `developer-agent`, `code-reviewer`
- Keep names descriptive and unique within the pipeline
- Document the pipeline structure so all agents know their neighbors

## Best Practices

1. **Use consistent agent names** - Each agent should use a unique, consistent name
2. **Always acknowledge messages** - Call `ack_message` after receiving to confirm delivery
3. **Include context in messages** - Provide enough detail for the receiving agent to understand
4. **Handle blocking calls** - `get_previous_commit` and `get_next_feedback` block until messages arrive
5. **Use structured content** - Consider using JSON or markdown in message content for clarity
6. **Define identity in system prompts** - Include agent name and pipeline neighbors in the system prompt

## Error Handling

The tools return error messages in these cases:

| Error | Cause |
|-------|-------|
| `"from, to, and content are required"` | Missing required parameters |
| `"agent_name is required"` | Missing agent name parameter |
| `"message_id is required"` | Missing message ID for acknowledgment |
| `"message not found"` | Invalid message ID for acknowledgment |
| `"message queue full for target agent"` | Target agent's queue is full (100 messages max) |
| `"context cancelled while waiting"` | Request timed out while waiting |

## Running the Server

### Using Docker

```bash
# Build the Docker image
docker build -t wlgoagentscoopmcp .

# Run the container
docker run -d -p 3000:3000 -p 3001:3001 -v .:/app -w /app --name wlgoagentscoopmcp wlgoagentscoopmcp

# Start the server inside container
docker exec -d wlgoagentscoopmcp bash -c "go run main.go"
```

### Running Directly

```bash
# Ensure .env file exists (copy from .env.example if needed)
cp .env.example .env

# Run the server
go run main.go
```

The MCP server listens on port 3001 by default (configurable via `MCP_PORT` env var).
