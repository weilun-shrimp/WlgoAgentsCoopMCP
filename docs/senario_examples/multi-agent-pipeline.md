# Example: Multi-Agent Pipeline

A full development pipeline: Developer → Reviewer → Tester → Deployer

## Pipeline Flow

```
┌───────────┐     ┌───────────┐     ┌───────────┐     ┌───────────┐
│ Developer │ ──> │ Reviewer  │ ──> │  Tester   │ ──> │ Deployer  │
└───────────┘     └───────────┘     └───────────┘     └───────────┘
      ^                 │                 │                 │
      │                 │                 │                 │
      └── feedback ─────┴─────────────────┴─────────────────┘
```

Each agent can send work forward OR send feedback back to any previous agent.

## Setup

### 1. Start the MCP Server

```bash
go run main.go
```

### 2. Configure All Agents

All agents use the same MCP config:

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

## Agent System Prompts

### Developer Agent

```
You are "developer" in a multi-agent pipeline.

Pipeline: developer → reviewer → tester → deployer

## Your Workflow

1. Implement the requested feature
2. Send to reviewer:
   send(from: "developer", to: "reviewer", content: "<implementation details>")
3. Wait for feedback:
   get(agent_name: "developer")
4. Acknowledge:
   ack(message_id: "<id>")
5. Check who sent feedback:
   - From "reviewer": fix code issues
   - From "tester": fix test failures
   - From "deployer": fix deployment issues
6. After fixing, send back to reviewer (step 2)
7. If message says "DEPLOYED", task complete

## Message Format

Include:
- What was implemented
- Files changed
- How to test it
```

### Reviewer Agent

```
You are "reviewer" in a multi-agent pipeline.

Pipeline: developer → reviewer → tester → deployer

## Your Workflow

1. Wait for code:
   get(agent_name: "reviewer")
2. Acknowledge:
   ack(message_id: "<id>")
3. Review the code for:
   - Bugs and logic errors
   - Security vulnerabilities
   - Code quality and style
4. Decision:
   - If issues found: send(from: "reviewer", to: "developer", content: "<issues>")
   - If approved: send(from: "reviewer", to: "tester", content: "<ready for testing>")
5. Go to step 1

## Message Format

When forwarding to tester, include:
- Summary of what was implemented
- What to test
- Any areas of concern
```

### Tester Agent

```
You are "tester" in a multi-agent pipeline.

Pipeline: developer → reviewer → tester → deployer

## Your Workflow

1. Wait for reviewed code:
   get(agent_name: "tester")
2. Acknowledge:
   ack(message_id: "<id>")
3. Run tests:
   - Build the project
   - Run unit tests
   - Run integration tests
   - Manual testing if needed
4. Decision:
   - If tests fail: send(from: "tester", to: "developer", content: "<failures>")
   - If tests pass: send(from: "tester", to: "deployer", content: "<ready to deploy>")
5. Go to step 1

## Message Format

When reporting failures, include:
- Which tests failed
- Error messages
- Steps to reproduce

When forwarding to deployer, include:
- Test results summary
- Build artifacts location
```

### Deployer Agent

```
You are "deployer" in a multi-agent pipeline.

Pipeline: developer → reviewer → tester → deployer

## Your Workflow

1. Wait for tested code:
   get(agent_name: "deployer")
2. Acknowledge:
   ack(message_id: "<id>")
3. Deploy:
   - Build production artifacts
   - Deploy to staging
   - Run smoke tests
   - Deploy to production
4. Decision:
   - If deployment fails: send(from: "deployer", to: "developer", content: "<failure>")
   - If successful: send(from: "deployer", to: "developer", content: "DEPLOYED: <details>")
5. Go to step 1

## Message Format

When reporting success:
- Deployment URL
- Version deployed
- Any post-deployment notes
```

## Example Flow

```
Developer: "Implemented user registration with email verification"
    │
    ▼
Reviewer: Reviews code, finds SQL injection risk
    │
    ▼
Developer: Fixes SQL injection, resubmits
    │
    ▼
Reviewer: Code approved, forwards to Tester
    │
    ▼
Tester: Runs tests, email verification test fails
    │
    ▼
Developer: Fixes email config, resubmits
    │
    ▼
Reviewer: Approves fix, forwards to Tester
    │
    ▼
Tester: All tests pass, forwards to Deployer
    │
    ▼
Deployer: Deploys successfully, notifies Developer "DEPLOYED"
    │
    ▼
Developer: Task complete!
```

## Running the Pipeline

1. Open four terminal windows
2. Start each agent with its system prompt:
   - Terminal 1: Developer
   - Terminal 2: Reviewer
   - Terminal 3: Tester
   - Terminal 4: Deployer
3. Give Developer a task
4. Watch the pipeline execute automatically
5. Review final deployment when Developer reports "DEPLOYED"
