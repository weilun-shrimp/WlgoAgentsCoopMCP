# Example: Task Manager Architecture

The highest level of automation: Human talks only to a Manager agent, who acts as a **planner** - clarifying requirements, designing the execution plan, and orchestrating all worker agents.

## Architecture

```
+─────────────────────────────────────────────────────────────+
|                         HUMAN                               |
|              (Only interacts with Manager)                  |
+─────────────────────────────+───────────────────────────────+
                              |
                              v
+─────────────────────────────────────────────────────────────+
|                     MANAGER AGENT                           |
|                  (Planner & Orchestrator)                   |
|                                                             |
|   Phase 1: CLARIFY  →  Phase 2: PLAN  →  Phase 3: EXECUTE   |
+─────────────────────────────+───────────────────────────────+
                              |
          +───────────────────+───────────────────+
          |                   |                   |
          v                   v                   v
    +-----------+       +-----------+       +-----------+
    | Developer |       | Reviewer  |       |  Tester   |
    |   Agent   |       |   Agent   |       |   Agent   |
    +-----------+       +-----------+       +-----------+
          |                   |                   |
          +───────────────────+───────────────────+
                              |
                              v
                        +-----------+
                        | Deployer  |
                        |   Agent   |
                        +-----------+
```

---

## Manager Agent: The Three Phases

The Manager agent operates as a planner with three distinct phases. It never starts work until requirements are crystal clear and the human has approved the plan.

### Phase 1: CLARIFY

**Goal**: Achieve 100% clarity on what the human wants.

The Manager:
1. Analyzes the initial request from the human
2. Identifies ambiguities, missing information, or assumptions
3. Asks targeted clarifying questions
4. Continues the dialogue until every detail is understood
5. Summarizes understanding and confirms with human

**Key Principle**: Never assume. Always ask.

### Phase 2: PLAN

**Goal**: Create a detailed, actionable execution plan.

The Manager:
1. Breaks down the requirement into discrete tasks
2. Assigns each task to the appropriate agent based on expertise
3. Defines execution order and dependencies between tasks
4. Specifies success criteria for each task
5. Identifies potential risks and contingencies
6. Presents the complete plan to the human for approval

**Key Principle**: Human approves the plan before any work begins.

### Phase 3: EXECUTE

**Goal**: Orchestrate agents to complete the plan.

The Manager:
1. Dispatches tasks to agents in the planned order
2. Monitors progress and collects results
3. Handles failures and re-plans if necessary
4. Coordinates handoffs between agents
5. Reports final result to the human

**Key Principle**: Human only sees the final result (unless intervention is needed).

---

## Complete Example Session

### Scenario: Human wants to build a user authentication system

---

### Phase 1: CLARIFY

**Human initiates the request:**

```
Human → Manager:
"Build a user authentication system for my web app"
```

**Manager analyzes and asks clarifying questions:**

```
Manager → Human:
"I'll help you build that. Before I create a plan, I need to clarify several points:

AUTHENTICATION METHODS:
1. Which login methods do you need?
   - Email/password only?
   - Social login (Google, GitHub, etc.)?
   - Magic link (passwordless)?

FEATURES:
2. Which features should be included?
   - User registration
   - Login/logout
   - Password reset
   - Email verification
   - Remember me / session management

SECURITY:
3. What security measures are required?
   - Rate limiting on login attempts?
   - Two-factor authentication (2FA)?
   - Account lockout after failed attempts?

TECHNICAL:
4. What's your current tech stack?
   - Backend framework?
   - Database?
   - Frontend framework?

5. Where will this be deployed?
   - Cloud provider?
   - Container/serverless?"
```

**Human provides answers:**

```
Human → Manager:
"- Email/password only for now
- Need: registration, login, logout, password reset
- Yes to rate limiting, no 2FA for now
- Backend: Go with Gin, Database: PostgreSQL, Frontend: React
- Deploy to AWS ECS with Docker"
```

**Manager confirms understanding:**

```
Manager → Human:
"Let me confirm my understanding:

BUILD: User authentication system
- Auth method: Email/password
- Features: Registration, login, logout, password reset
- Security: Rate limiting (no 2FA)
- Stack: Go/Gin + PostgreSQL + React
- Deploy: AWS ECS with Docker

Is this correct? Any additions?"
```

```
Human → Manager:
"Yes, that's correct. Please proceed."
```

---

### Phase 2: PLAN

**Manager creates and presents the execution plan:**

```
Manager → Human:
"Here's my execution plan:

═══════════════════════════════════════════════════════════════
                    EXECUTION PLAN
═══════════════════════════════════════════════════════════════

TASK 1: Backend Authentication API
  Agent: Developer
  Work:
    - Create user model and database migrations
    - Implement registration endpoint with password hashing
    - Implement login endpoint with JWT token generation
    - Implement logout endpoint
    - Implement password reset flow
    - Add rate limiting middleware
  Output: Go API code + SQL migrations

TASK 2: Backend Code Review
  Agent: Reviewer
  Depends on: Task 1
  Work:
    - Review code for security vulnerabilities
    - Check password hashing implementation
    - Verify JWT handling
    - Review rate limiting logic
    - Check SQL injection prevention
  Output: Review report + required fixes

TASK 3: Apply Review Fixes
  Agent: Developer
  Depends on: Task 2
  Work:
    - Address all issues from code review
  Output: Updated code

TASK 4: Frontend Authentication UI
  Agent: Developer
  Depends on: Task 3
  Work:
    - Create login page component
    - Create registration page component
    - Create password reset flow
    - Implement JWT token storage and refresh
    - Add protected route handling
  Output: React components

TASK 5: Frontend Code Review
  Agent: Reviewer
  Depends on: Task 4
  Work:
    - Review for XSS vulnerabilities
    - Check token storage security
    - Review form validation
  Output: Review report + required fixes

TASK 6: Apply Frontend Fixes
  Agent: Developer
  Depends on: Task 5
  Work:
    - Address all issues from code review
  Output: Updated code

TASK 7: Testing
  Agent: Tester
  Depends on: Task 6
  Work:
    - Write and run unit tests for auth endpoints
    - Write and run integration tests
    - Test rate limiting behavior
    - Test password reset flow end-to-end
    - Security testing (OWASP checklist)
  Output: Test results + coverage report

TASK 8: Fix Test Failures
  Agent: Developer
  Depends on: Task 7 (if failures)
  Work:
    - Fix any failing tests
  Output: All tests passing

TASK 9: Deployment
  Agent: Deployer
  Depends on: Task 8
  Work:
    - Create Dockerfile for backend
    - Create Docker Compose for local dev
    - Create AWS ECS task definition
    - Set up environment variables
    - Deploy to staging environment
  Output: Deployed application

TASK 10: Deployment Verification
  Agent: Tester
  Depends on: Task 9
  Work:
    - Smoke test deployed application
    - Verify all endpoints work
  Output: Verification report

═══════════════════════════════════════════════════════════════

DEPENDENCY GRAPH:

  [Task 1] → [Task 2] → [Task 3] → [Task 4] → [Task 5]
                                                  ↓
  [Task 10] ← [Task 9] ← [Task 8] ← [Task 7] ← [Task 6]

═══════════════════════════════════════════════════════════════

Do you approve this plan?"
```

**Human approves:**

```
Human → Manager:
"Approved. Start execution."
```

---

### Phase 3: EXECUTE

**Manager orchestrates the agents:**

```
Manager → Developer:
"TASK 1: Create backend authentication API.

Requirements:
- Go with Gin framework
- PostgreSQL database
- Endpoints: POST /register, POST /login, POST /logout, POST /password-reset
- Use bcrypt for password hashing
- Use JWT for session tokens
- Add rate limiting: max 5 login attempts per minute per IP

Deliver: All Go code files + SQL migration files"
```

```
Developer → Manager:
"Task 1 complete. Created:
- /internal/models/user.go
- /internal/handlers/auth.go
- /internal/middleware/ratelimit.go
- /migrations/001_create_users.sql
- /migrations/002_create_password_resets.sql"
```

```
Manager → Reviewer:
"TASK 2: Review the authentication code.

Files to review:
- /internal/models/user.go
- /internal/handlers/auth.go
- /internal/middleware/ratelimit.go

Focus areas:
- Password hashing security
- JWT implementation
- SQL injection prevention
- Rate limiting effectiveness

Deliver: Review report with any required fixes"
```

```
Reviewer → Manager:
"Task 2 complete. Review findings:

CRITICAL:
- [auth.go:45] JWT secret is hardcoded. Move to environment variable.

MEDIUM:
- [auth.go:78] Add password complexity validation
- [ratelimit.go:23] Use Redis for rate limiting in production

LOW:
- [user.go:12] Add email format validation"
```

```
Manager → Developer:
"TASK 3: Fix the review issues.

CRITICAL (must fix):
- Move JWT secret to environment variable

MEDIUM (must fix):
- Add password complexity validation
- Prepare rate limiter for Redis (can use memory for now)

LOW (should fix):
- Add email format validation"
```

*...execution continues through all tasks...*

**Manager reports final result:**

```
Manager → Human:
"═══════════════════════════════════════════════════════════════
                    EXECUTION COMPLETE
═══════════════════════════════════════════════════════════════

STATUS: Success

DELIVERABLES:
✓ Backend API with authentication endpoints
✓ Frontend React components for auth flows
✓ 94% test coverage
✓ Deployed to AWS ECS staging environment

ENDPOINTS AVAILABLE:
- POST /api/register
- POST /api/login
- POST /api/logout
- POST /api/password-reset/request
- POST /api/password-reset/confirm

STAGING URL: https://staging.yourapp.aws.com

NEXT STEPS (your decision):
1. Test the staging environment
2. Approve for production deployment
3. Request changes

═══════════════════════════════════════════════════════════════"
```

---

## Ready-to-Use Prompts

All agent prompts are available in the [prompt_examples](../prompt_examples/) folder:

| Agent | Prompt File |
|-------|-------------|
| Manager | [manager_prompt.md](../prompt_examples/manager_prompt.md) |
| Developer | [developer_prompt.md](../prompt_examples/developer_prompt.md) |
| Reviewer | [reviewer_prompt.md](../prompt_examples/reviewer_prompt.md) |
| Tester | [tester_prompt.md](../prompt_examples/tester_prompt.md) |
| Deployer | [deployer_prompt.md](../prompt_examples/deployer_prompt.md) |

---

## MCP Tool Usage

### Sending a Task to an Agent

```json
{
  "tool": "send",
  "arguments": {
    "from": "manager",
    "to": "developer",
    "content": "TASK 1: Create backend authentication API\n\nRequirements:\n- Go with Gin framework\n- PostgreSQL database\n- JWT for session tokens\n- bcrypt for password hashing\n\nDeliverables: Go code files + SQL migrations"
  }
}
```

### Waiting for a Response

```json
{
  "tool": "get",
  "arguments": {
    "agent_name": "manager"
  }
}

// Response:
{
  "success": true,
  "message_id": "msg-abc123",
  "message": {
    "id": "msg-abc123",
    "from": "developer",
    "to": "manager",
    "content": "Task 1 complete. Created:\n- /internal/handlers/auth.go\n- /internal/models/user.go\n- /migrations/001_create_users.sql",
    "timestamp": 1699900000000
  }
}
```

### Acknowledging a Message

```json
{
  "tool": "ack",
  "arguments": {
    "message_id": "msg-abc123"
  }
}
```

---

## Benefits of This Architecture

| Benefit | Description |
|---------|-------------|
| **Human Focus** | Human only defines the goal, not the steps |
| **Clear Requirements** | Clarify phase prevents misunderstandings |
| **Approved Plans** | Human sees exactly what will happen before it starts |
| **Autonomous Execution** | No human intervention during work |
| **Quality Built-in** | Code review and testing are part of every plan |
| **Transparent Results** | Human gets a clear summary when done |

---

## When to Use This Pattern

**Use Task Manager when:**
- Tasks are complex with multiple steps
- Multiple agent types are needed
- Quality gates (review, testing) are required
- Human doesn't want to manage individual agents

**Use simpler patterns when:**
- Single agent can complete the task
- Task is straightforward with no ambiguity
- Human wants direct control over each step
