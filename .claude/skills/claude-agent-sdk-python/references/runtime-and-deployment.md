# Claude Agent SDK Python Runtime and Deployment Notes

Use this file when the user's question involves infrastructure, MCP behavior, rewind flows, or production deployment.

## MCP integration

### Permission model

- MCP tools follow `mcp__<server-name>__<tool-name>`.
- Without permission, Claude may see MCP tools but not call them.
- Common approval patterns:
  - exact tool names
  - server wildcards like `mcp__github__*`
  - broader permission modes when appropriate

### Transport choices

- stdio: local processes you spawn with `command` + `args`
- http / sse: hosted remote MCP endpoints
- SDK MCP server: tools defined directly inside your Python app

### Troubleshooting MCP

If tools are not being called:

- verify the server connected in the init metadata
- verify permission rules or `allowed_tools`
- verify auth env vars / headers are passed correctly

If the server status is `failed`:

- check missing env vars
- check command availability and PATH
- check connection strings or remote URLs
- consider startup timeout issues for slow servers

### MCP tool search

When many MCP tools are loaded, context cost can get large.

Important points:

- tool search can defer loading until tools are needed
- it depends on model support for tool references
- Haiku does not support it
- configure it through `ENABLE_TOOL_SEARCH` in `env`

## File checkpointing

### What it does

Checkpointing tracks file changes made through:

- `Write`
- `Edit`
- `NotebookEdit`

It does not track Bash-based file mutations.

### Python requirements

- set `enable_file_checkpointing=True`
- set `extra_args={"replay-user-messages": None}` if you need checkpoint UUIDs from `UserMessage.uuid`

### Rewind flow

Typical flow:

1. run the agent with checkpointing enabled
2. capture checkpoint UUIDs from `UserMessage.uuid`
3. capture `session_id` from `ResultMessage`
4. resume the session later with an empty prompt
5. call `await client.rewind_files(checkpoint_id)` on the live connection

### Common failure modes

- UUID missing: forgot `replay-user-messages`
- no checkpoint found: checkpointing was not enabled in the original session
- transport not ready for writing: tried to rewind after the response loop had already finished without reopening the session

## Hosting and production deployment

### Mental model

The Agent SDK is not just a stateless text API wrapper. It runs a persistent tool-using process with filesystem state, shell state, and session state.

### Baseline production guidance

- run inside sandboxed containers
- control filesystem and network access
- provision CPU, RAM, and disk intentionally
- monitor token cost, tool usage, and container health
- set explicit `max_turns` and permission strategy

### Common deployment patterns

- Ephemeral sessions: one task per container
- Long-running sessions: persistent background agents
- Hybrid sessions: restore state/history into short-lived containers
- Single shared container: niche; avoid unless agents truly need shared state

### Good advice to give users

- prefer isolated environments for `bypassPermissions`
- do not assume long-running agents will stop on their own; configure limits
- persist app-level state explicitly when cross-host session resume is fragile
- distinguish session persistence from file rollback and from process lifetime
