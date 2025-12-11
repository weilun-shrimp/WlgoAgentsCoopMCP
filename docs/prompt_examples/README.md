# Agent Prompt Examples

Ready-to-use system prompts for multi-agent cooperation.

## Available Prompts

| File | Agent | Description |
|------|-------|-------------|
| [manager_prompt.md](manager_prompt.md) | Manager | Planner & orchestrator - talks to human, coordinates all other agents |
| [developer_prompt.md](developer_prompt.md) | Developer | Writes code based on task assignments |
| [reviewer_prompt.md](reviewer_prompt.md) | Reviewer | Reviews code for quality and security |
| [tester_prompt.md](tester_prompt.md) | Tester | Writes and runs tests |
| [deployer_prompt.md](deployer_prompt.md) | Deployer | Handles deployment and infrastructure |

## How to Use

1. **Choose your architecture**:
   - **Two agents**: Developer + Reviewer (see [two-agents.md](../examples/two-agents.md))
   - **Multiple agents**: Developer + Reviewer + Tester + Deployer (see [multi-agent.md](../examples/multi-agent.md))
   - **Task Manager**: Manager + all worker agents (see [task-manager.md](../examples/task-manager.md))

2. **Copy the prompt** from the relevant file

3. **Configure your AI agent** with the prompt as the system message

4. **Connect to MCP server** - ensure each agent has access to the MCP tools (`send`, `get`, `ack`)

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     TASK MANAGER MODE                       │
│                                                             │
│   Human ←→ Manager ←→ Developer/Reviewer/Tester/Deployer    │
│                                                             │
│   Human only talks to Manager                               │
│   Manager orchestrates all work                             │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    DIRECT AGENT MODE                        │
│                                                             │
│   Human ←→ Developer ←→ Reviewer                            │
│                                                             │
│   Human talks directly to Developer                         │
│   Developer coordinates with Reviewer                       │
└─────────────────────────────────────────────────────────────┘
```

## MCP Tools Reference

All agents use these three tools:

| Tool | Purpose | Arguments |
|------|---------|-----------|
| `send` | Send message to another agent | `from`, `to`, `content` |
| `get` | Wait for and receive a message | `agent_name` |
| `ack` | Acknowledge and remove a message | `message_id` |

