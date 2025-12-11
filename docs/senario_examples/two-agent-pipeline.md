# Example: Two-Agent Pipeline

A simple Developer + Reviewer pipeline where code is written, reviewed, and iterated until approved.

## Pipeline Flow

```
┌────────────┐                      ┌────────────┐
│ Developer  │ ───── code ─────────>│  Reviewer  │
│            │                      │            │
│            │ <── feedback/approve │            │
└────────────┘                      └────────────┘
      │                                   │
      └──── iterate until approved ───────┘
```

## Setup

### 1. Start the MCP Server

```bash
go run main.go
```

### 2. Configure Both Agents

Both agents need the MCP server in their config:

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

### 3. Agent System Prompts

**Developer Agent Prompt:**

```
You are "developer" in a two-agent code review pipeline.

Your partner is "reviewer" who will review your code.

## Your Workflow

1. Read the task requirements
2. Implement the solution
3. Send your work to reviewer:
   send(from: "developer", to: "reviewer", content: "<describe what you implemented and which files>")
4. Wait for feedback:
   get(agent_name: "developer")
5. After receiving, acknowledge:
   ack(message_id: "<id from the message>")
6. If feedback requests changes, make them and go to step 3
7. If feedback says "APPROVED", you're done

## Message Format

When sending to reviewer, include:
- What you implemented
- Which files were changed
- Any design decisions you made
```

**Reviewer Agent Prompt:**

```
You are "reviewer" in a two-agent code review pipeline.

Your partner is "developer" who sends you code to review.

## Your Workflow

1. Wait for developer's submission:
   get(agent_name: "reviewer")
2. Acknowledge the message:
   ack(message_id: "<id from the message>")
3. Review the code:
   - Check for bugs
   - Check for security issues
   - Check code quality
   - Run tests if applicable
4. Send feedback:
   send(from: "reviewer", to: "developer", content: "<your feedback>")
5. Go to step 1

## Feedback Format

If changes needed:
- List specific issues
- Suggest how to fix them

If everything passes:
- Say "APPROVED" clearly
- Summarize what was good
```

## Example Conversation

**Developer sends:**
```
Implemented user login feature.

Files changed:
- src/auth/login.go (new)
- src/handlers/auth_handler.go (modified)

Design decisions:
- Used bcrypt for password hashing
- JWT tokens expire in 24 hours
```

**Reviewer receives, reviews, sends back:**
```
Review feedback:

Issues found:
1. Missing rate limiting on login endpoint - could allow brute force
2. JWT secret is hardcoded in auth_handler.go line 15

Please fix these and resubmit.
```

**Developer fixes and sends again:**
```
Fixed the issues:

1. Added rate limiting middleware (max 5 attempts per minute)
2. Moved JWT secret to environment variable

Files changed:
- src/middleware/rate_limit.go (new)
- src/handlers/auth_handler.go (modified)
```

**Reviewer approves:**
```
APPROVED

All issues resolved. Good implementation:
- Rate limiting properly configured
- Secrets now loaded from environment
- Code is clean and well-structured
```

## Running the Pipeline

1. Open two terminal windows
2. Start Developer agent in terminal 1 with its system prompt
3. Start Reviewer agent in terminal 2 with its system prompt
4. Give Developer a task (e.g., "Implement user registration")
5. Watch them collaborate automatically
6. Review the final result when Developer reports completion
