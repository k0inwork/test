# Jules Bridge Daemon: Agentic Session Control

## Overview
The `jules-bridge-daemon` is a standalone, general-purpose tool designed to provide external visibility and control over an AI agent's (Jules) execution session. It acts as a bidirectional synchronization layer between the agent's internal state (plans, steps, thoughts) and an external control script or dashboard.

## Architecture
The system consists of three main components:
1. **The Agent (Jules):** The AI performing the task. It uses local tools to communicate with the Bridge Daemon.
2. **The Bridge Daemon:** A background process (daemon) running within the Jules VM. It manages network communication with the external world.
3. **The External Controller:** A script or service that receives feedback from Jules and sends back control commands.

### Communication Flow
1. **Feedback (Outbound):** Jules reports its current plan or status to the Bridge Daemon. The Daemon `POST`s this data to a configured `FEEDBACK_ENDPOINT`.
2. **Control (Inbound):** After reporting, Jules enters a "Waiting" state. The Daemon listens for a "Hook" from the External Controller. This hook provides the next instruction (Continue, Edit Plan, Stop).

## Plan State Machine
The daemon tracks the lifecycle of an agent's task:
- `DRAFT`: The plan is created but not yet approved.
- `APPROVED`: The controller has given the green light to proceed.
- `EXECUTING`: A step is currently being performed.
- `PAUSED`: Execution is temporarily halted for manual intervention or clarification.
- `COMPLETED`: The task is finished.
- `FAILED`: An unrecoverable error occurred.

## Integration with Jules VM
- **Lifecycle:** The daemon is configured as a `systemd` service or added to the VM's startup sequence (`.profile` or similar).
- **Hardcoded Snapshots:** The daemon and its configuration become part of the VM's base image, ensuring every new session is "hookable" by default.

## API & Data Structures

### Outbound Feedback (JSON)
```json
{
  "session_id": "uuid-1234",
  "timestamp": "2023-10-27T10:00:00Z",
  "event": "PLAN_CREATED",
  "payload": {
    "steps": [
      {"id": 1, "description": "Research codebase"},
      {"id": 2, "description": "Modify main.go"}
    ]
  }
}
```

### Inbound Control (JSON)
```json
{
  "command": "CONTINUE",
  "modified_plan": null,
  "reason": "Plan looks good"
}
```

## Security & Connectivity
- **Webhooks/Polling:** To bypass firewalls/NAT, the daemon can use long-polling or a persistent WebSocket connection to a central relay.
- **Identity:** Each VM session reports a unique `JULES_SESSION_ID` to allow the Controller to distinguish between multiple running agents.
