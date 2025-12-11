# Quick Start

Get multi-agent cooperation running in 5 minutes.

## 1. Start the MCP Server

**Option A: Docker**
```bash
docker build -t wlgo-agents-coop-mcp .
docker run -d -p 3000:3000 -w /app -v .:/app --name wlgo-agents-coop-mcp wlgo-agents-coop-mcp
docker exec -d wlgo-agents-coop-mcp bash -c "go run main.go"
```

**Option B: Go directly**
```bash
go run main.go
```

Server runs at `ws://localhost:3000/mcp` (WebSocket)

## 2. Configure Your AI Tool

Add MCP server to your AI CLI tool config (e.g., Claude Code):

```json
{
  "mcpServers": {
    "agents-coop": {
      "type": "websocket",
      "url": "ws://localhost:3000/mcp"
    }
  }
}
```

## 3. Set Up Permissions

Create `.claude/settings.json` in your project to allow commands without prompts:

```json
{
  "permissions": {
    "allow": [
      "mcp__WlgoAgentsCoopMCP__get_previous_commit",
      "mcp__WlgoAgentsCoopMCP__commit_to_next",
      "mcp__WlgoAgentsCoopMCP__get_next_feedback",
      "mcp__WlgoAgentsCoopMCP__ack_message",
      "mcp__WlgoAgentsCoopMCP__feedback_to_previous",
      "WebFetch(domain:github.com)",
      "Bash(mkdir:*)",
      .... Or anything you want
    ]
  }
}
```

## 4. Start Agents

Open multiple terminals, each running your AI CLI.

**Terminal 1 - Manager Agent**

Paste the [manager prompt](prompt_examples/manager_prompt.md) directly to the agent.

**Terminal 2 - Developer Agent**

Paste the [developer prompt](prompt_examples/developer_prompt.md) directly to the agent.

**Terminal 3 - Reviewer Agent** (optional)

Paste the [reviewer prompt](prompt_examples/reviewer_prompt.md) directly to the agent.

## 5. Give Task to Manager

Talk to the Manager agent with your goal:

```
Build a Go HTTP service with Fiber framework.
- Dockerfile with Ubuntu 24.04 and Go 1.25
- .env file with LISTEN_PORT and APP_NAME
- Graceful shutdown
- CRUD API for posts (in-memory storage)
- README with docker build/run instructions
```

The Manager will:
1. Ask clarifying questions
2. Create an execution plan
3. Get your approval
4. Dispatch tasks to Developer/Reviewer
5. Report final results

## Available Prompts

| Agent | File |
|-------|------|
| Manager | [manager_prompt.md](prompt_examples/manager_prompt.md) |
| Developer | [developer_prompt.md](prompt_examples/developer_prompt.md) |
| Reviewer | [reviewer_prompt.md](prompt_examples/reviewer_prompt.md) |
| Tester | [tester_prompt.md](prompt_examples/tester_prompt.md) |
| Deployer | [deployer_prompt.md](prompt_examples/deployer_prompt.md) |

## Next Steps

- [Two-Agent Example](examples/two-agents.md) - Developer + Reviewer
- [Multi-Agent Example](examples/multi-agent.md) - Full pipeline
- [Task Manager Example](examples/task-manager.md) - Manager orchestration
