# WlgoAgentsCoopMCP

## Problem

Current AI agents work in isolation. Each agent handles a single task and requires human intervention to coordinate with other agents. This creates bottlenecks:

- Humans must manually pass outputs between agents
- Humans must review intermediate results at every step
- Complex tasks get fragmented into small, disconnected pieces
- The human becomes the communication layer between AI systems

## Goal

Enable AI agents to collaborate autonomously on high-level tasks. Humans should only need to:

1. Define the goal
2. Review the final result

Everything in between — coordination, iteration, feedback loops — should happen agent-to-agent without human interrupt.

## How It Works

WlgoAgentsCoopMCP is an MCP server that provides three simple tools for agent-to-agent messaging:

| Tool   | Description                              |
|--------|------------------------------------------|
| `send` | Send a message to another agent          |
| `get`  | Wait and receive a message               |
| `ack`  | Acknowledge and remove message from storage |

```
┌─────────────┐                      ┌─────────────┐
│   Agent A   │  send(to: "B", ...)  │   Agent B   │
│             │ ────────────────────>│             │
│             │                      │  get("B")   │
│             │                      │  ack(msg)   │
│             │                      │             │
│             │  send(to: "A", ...)  │             │
│             │ <────────────────────│             │
│  get("A")   │                      │             │
│  ack(msg)   │                      │             │
└─────────────┘                      └─────────────┘
```

## Quick Start

### 1. Deploy the Server

**Option A: Using Docker (Recommended)**

```bash
# Build the image
docker build -t wlgoagentscoopmcp .

# Run the container
docker run -d -p 3000:3000 -w /app -v .:/app --name wlgoagentscoopmcp wlgoagentscoopmcp

# Start the server
docker exec -d wlgoagentscoopmcp bash -c "go run main.go"

# Verify it's running
curl http://localhost:3000/ping
```

**Option B: Run Directly**

```bash
# Install dependencies
go mod download

# Run the server
go run main.go
```

The MCP server runs on HTTP at `http://localhost:3000/mcp`.

### 2. Register MCP Server to Your AI CLI Tool

Add to your MCP configuration file:

**Claude Code CLI** (`.mcp.json` or `~/.claude.json`):

```json
{
  "mcpServers": {
    "agents-coop": {
      "type": "http",
      "url": "http://localhost:3000/mcp"
    }
  }
}
```

### 3. Teach Your AI Agents to Use the Tools

Add instructions to your agent's system prompt:

```
You are "developer" in a multi-agent system.

To communicate with other agents, use these MCP tools:
- send(from: "developer", to: "<agent-name>", content: "<message>") - Send a message
- get(agent_name: "developer") - Wait for incoming messages (blocks until received)
- ack(message_id: "<id>") - Acknowledge and remove the message after processing
```

## Documentation

- [Quick Start](docs/quick-start.md) - Get running in 5 minutes
- [Agent Communication Guide](docs/agent-guide.md) - Detailed tool reference and examples
- [Example: Two-Agent Pipeline](docs/examples/two-agent-pipeline.md) - Developer + Reviewer
- [Example: Multi-Agent Pipeline](docs/examples/multi-agent-pipeline.md) - Dev → Review → Test → Deploy
- [Example: Task Manager](docs/examples/task-manager.md) - Human → Manager → Multiple Agents

## Environment

| Component     | Version      |
|---------------|--------------|
| Go            | 1.25         |
| Base Image    | Ubuntu 24.04 |
| Web Framework | Fiber v2.52.6|
| MCP Transport | HTTP         |
