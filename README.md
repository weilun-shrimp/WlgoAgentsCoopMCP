# WlgoAgentsCoopMCP

## Overview
**WlgoAgentsCoopMCP** (Wlgo Agents Cooperative Multi-Agent Communication Pipeline) is a server that allows multiple AI agents to connect, read task goals and their role-specific prompts, and collaboratively complete tasks in a structured, automated workflow. Tasks only start processing when the human operator triggers the **task start API**, giving full control over execution while supporting autonomous agent collaboration.

## Features
- **Controlled execution:** Agents wait for a task start signal before processing.  
- **Flexible pipeline:** Supports multiple stages and any type of agent roles.  
- **Autonomous collaboration:** Agents communicate, iterate, and progress tasks without human intervention.  
- **Scalable and extensible:** Multiple agents and multiple tasks can run concurrently.  
- **Transparent logging:** Full auditability of messages, outputs, and rollback.

## System Workflow

1. **Task Creation**
   - A human operator creates a task and defines its goal.  
   - Agents are assigned to the task with role-specific prompts.  
   - The task is stored on the MCP server in a **waiting state**.

2. **Agent Connection**
   - Agents connect to the MCP server and read the task goal and their prompts.  
   - Agents remain in a ready/waiting state until the task is started.

3. **Task Start**
   - The operator triggers the task using the **task start API**.  
   - The MCP server signals all connected agents to begin processing.  

4. **Pipeline Processing**
   - Tasks move through a **multi-stage pipeline**, where each agent performs its function.  
   - Agents communicate structured messages including outputs, approvals, and rollback.  
   - Failed outputs are sent back upstream for revision, creating an **iterative rollback loop**.  
   - Tasks continue iterating until all stages pass successfully.

5. **Completion and Logging**
   - Once all stages pass, the task is marked as complete.  
   - All communications, outputs, and rollback are logged for traceability.



## Visualize Agents cooperation pipeline flow
```
+-----------+     +-----------+ ---Commit---> +-----------+ ---Commit---> +-----------+     +-----------+
|   Start   | --> |   Agent1  |    Message    |   Agent2  |    Message    |   Agent3  | --> |  Complete |
+-----------+     +-----------+ <--Rollback-- +-----------+ <--Rollback-- +-----------+     +-----------+
```

Example
```
+-----------+     +-----------+ ---Commit---> +-----------+ ---Commit---> +-----------+     +-----------+
|   Start   | --> | Developer |    Message    |  Reviewer |    Message    |  Testing  | --> |  Complete |
+-----------+     +-----------+ <--Rollback-- +-----------+ <--Rollback-- +-----------+     +-----------+
```