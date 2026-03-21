---
name: claude-agent-sdk-python
description: Use this skill whenever the user wants to build, debug, refactor, or explain Python code that uses Anthropic's Claude Agent SDK. Trigger even when the user mentions `claude_agent_sdk`, `query()`, `ClaudeSDKClient`, `ClaudeAgentOptions`, hooks, `can_use_tool`, `AskUserQuestion`, `PermissionResultAllow`, `PermissionResultDeny`, `tools=`, sessions, `resume`, `fork_session`, subagents, `AgentDefinition`, MCP servers, `allowed_tools`, `disallowed_tools`, `permission_mode`, `setting_sources`, `ResultMessage`, structured output, `rewind_files`, file checkpointing, or migrating from the older Claude Code SDK. Use it especially for worker-style agents, custom in-process tools, permission debugging, interactive approval flows, session-aware agents, MCP-heavy applications, and Python applications that need Claude Code capabilities without shelling out manually.
---

# Claude Agent SDK Python

Use this skill to produce accurate, current guidance for Anthropic's Python Claude Agent SDK instead of falling back to generic Claude API advice.

## Compatibility

This skill assumes `claude-agent-sdk` is installed and the bundled or configured Claude Code CLI can run.

If the runtime setup is missing, say so clearly and avoid pretending examples were executed.

## What this skill is for

Reach for this skill when the user needs help with:
- one-shot or bounded agent runs with `query()`
- interactive or multi-turn sessions with `ClaudeSDKClient`
- streaming input versus single-message input
- sessions, `continue_conversation`, `resume`, and `fork_session`
- subagents with `AgentDefinition` and the `Agent` tool
- `ClaudeAgentOptions` setup
- structured output via `output_format`
- `AskUserQuestion` and approval flows through `can_use_tool`
- external MCP server wiring through `mcp_servers`
- MCP transport choices, wildcard permissions, and MCP tool search
- in-process SDK tools built with `@tool` and `create_sdk_mcp_server`
- hooks, permission behavior, and tool approval gotchas
- file checkpointing with `enable_file_checkpointing` and `rewind_files()`
- agent loop behavior, result subtypes, turn and budget limits, and effort tuning
- hosting and deployment tradeoffs for SDK-based agents
- upgrading old `Claude Code SDK` code to the newer `Claude Agent SDK`

## Grounding workflow

1. If the user gives a URL or names official docs, ground the answer in Anthropic's current documentation first.
2. Prefer exact SDK names and current behavior over approximate memory.
3. Call out version-sensitive behavior when it matters, especially around permissions and hooks.
4. If the user is already inside a repo, inspect existing SDK usage before proposing a new pattern.

## Choose the right SDK surface

### Use `query()` when

Use `query()` for simpler agent runs where the user wants one prompt processed with streaming responses and optional structured output.

This is usually the best fit for:
- background workers
- one task per request
- jobs that end with a single structured result
- code that does not need custom Python tools or long-lived interactive state

If the user needs real interactive approvals, streamed user messages, or a persistent in-process conversation object, consider streaming input or `ClaudeSDKClient` instead of a plain string prompt.

### Use `ClaudeSDKClient` when

Use `ClaudeSDKClient` when the user needs:
- a longer-lived conversation
- multiple `await client.query(...)` calls in one session
- custom in-process tools
- hooks defined as Python callbacks
- finer control over interactive behavior

If the user asks for hooks or Python-defined tools, default to `ClaudeSDKClient` unless there is a strong reason not to.

## Session and input-mode rules

- In Python, `ClaudeSDKClient` is the easiest way to keep one session alive across multiple prompts in the same process.
- For `query()`, use `resume` when the caller tracks a specific session ID and `fork_session=True` when they want a branch of an existing session.
- Session history persists automatically to disk; filesystem state does not magically branch with it.
- If `resume` appears not to work, check for a `cwd` mismatch before blaming the SDK.
- Streaming input is the recommended mode for interactive applications. Single-message input is fine for one-shot or worker flows.
- If the user needs image attachments, queued follow-up messages, or mid-flight interaction, steer them toward streaming input.

## Core implementation rules

### Installation and runtime

- The package is installed with `pip install claude-agent-sdk`.
- Python 3.10+ is required.
- The Claude Code CLI is bundled by default, so a separate CLI install is usually unnecessary.
- If the user needs a specific CLI binary, set `ClaudeAgentOptions(cli_path=...)`.

### Worker-style `query()` pattern

For worker or job-processing code:
1. Build a `ClaudeAgentOptions` object.
2. Put configuration such as `system_prompt`, `cwd`, `env`, `mcp_servers`, `output_format`, and `max_turns` there.
3. Iterate over `query(prompt=..., options=options)` asynchronously.
4. Read the final structured result from `ResultMessage.structured_output` when `output_format` is in use.
5. If the output may come back as a dict, model, or stringified JSON, normalize it carefully instead of assuming one shape.

### `ClaudeSDKClient` pattern

For interactive sessions:
1. Create `ClaudeAgentOptions`.
2. Open `async with ClaudeSDKClient(options=options) as client:`.
3. Send a prompt with `await client.query(...)`.
4. Stream the response with `async for message in client.receive_response():`.
5. Reuse the same client only when conversation continuity is actually useful.

## MCP guidance

### External MCP servers

When the user already has an MCP server process or wants to connect to one over stdio, configure it under `mcp_servers` with fields like:
- `type`
- `command`
- `args`
- `env`

Use this for tools such as Neo4j or other existing MCP processes.

### In-process SDK tools

When the tool logic naturally belongs inside the Python app, prefer SDK-defined tools:
- decorate a Python function with `@tool`
- wrap it with `create_sdk_mcp_server(...)`
- mount it via `ClaudeAgentOptions(mcp_servers={...})`

Prefer in-process tools over external MCP subprocesses when the user wants:
- simple business logic
- less deployment complexity
- type-hinted Python code
- faster local execution

## Permission and tool rules that people often get wrong

- `tools` controls the base built-in tool set Claude sees. It can be a list like `['Read', 'Grep']`, an empty list to disable built-in tools, or the preset `{"type": "preset", "preset": "claude_code"}`.
- `allowed_tools` is an auto-approval allowlist, not a hard capability allowlist.
- `disallowed_tools` is the right way to block tools.
- `can_use_tool` is the main programmatic permission callback when the user wants runtime allow/deny logic in Python code.
- In Python, `dontAsk` is not available as a permission mode.
- `bypassPermissions` approves everything that is not denied earlier by hooks or deny rules.
- `acceptEdits` auto-approves file edits and filesystem operations, not every tool.
- `can_use_tool` requires streaming mode. Do not pair it with a plain string prompt and expect it to work.
- In Python, approval flows that rely on `can_use_tool` also need the documented workaround hook that returns `{"continue_": True}` to keep the stream open.
- SDK MCP tools usually need an explicit approval strategy such as `allowed_tools`, `can_use_tool`, or hooks. Otherwise Claude may see the tool but fail to execute it.
- If the user restricts `tools` and still wants clarifying questions, include `AskUserQuestion` in that tool list.
- If the user wants tight control in Python, combine careful `allowed_tools`, `disallowed_tools`, hooks, and normal permission modes rather than pretending Python has a stricter mode than it really does.

## Hook rules and pitfalls

- Hook matchers match tool names or event filter values, not file paths.
- If filtering by file path, inspect `tool_input` inside the callback.
- To modify tool input, return `updatedInput` inside `hookSpecificOutput` and include `permissionDecision: "allow"`.
- Include the correct `hookEventName` in `hookSpecificOutput`.
- Return `{}` when the hook should allow the action without changes.
- Use async side effects only when the hook does not need to influence execution.

If the user wants SDK code to load shell-defined hooks or permission rules from `.claude/settings.json`, remember to include `setting_sources=["project"]`.

Also remember that no filesystem settings are loaded by default when `setting_sources` is omitted.

## Subagent rules and pitfalls

- The parent agent must have `Agent` in `allowed_tools` for Claude to invoke subagents.
- Write `AgentDefinition.description` as routing guidance: it tells Claude when to delegate.
- Do not give subagents the `Agent` tool. Subagents cannot spawn their own subagents.
- Subagents start with fresh conversation context. They do not inherit the parent's message history.
- If the user wants a subagent to preserve project instructions or filesystem-defined agents, remember `setting_sources`.
- On Windows, very long inline subagent prompts can hit command-line length limits. Prefer concise prompts or filesystem-based agents when prompts get bulky.

## Structured output rules and pitfalls

- Prefer Pydantic models in Python and pass `.model_json_schema()` into `output_format`.
- Keep schemas tight and realistic. Deep nesting plus many required fields increases failure risk.
- Handle `ResultMessage.subtype == "error_max_structured_output_retries"` explicitly.
- Make optional any field the task may legitimately fail to discover.

## MCP rules and pitfalls

- MCP tool names follow `mcp__<server-name>__<tool-name>`.
- Use wildcards like `mcp__github__*` when you really want all tools from one server.
- Check the init message or server status when MCP tools are not being called; many issues are connection or permission problems, not prompt problems.
- Use stdio for local server processes and HTTP/SSE for hosted MCP endpoints.
- Large MCP toolsets can bloat context. Mention MCP tool search when many servers or tools are involved.

## File checkpointing rules and pitfalls

- Checkpointing only tracks changes made through `Write`, `Edit`, and `NotebookEdit`.
- Bash-based file changes are not rewound.
- To receive checkpoint UUIDs in Python, enable `extra_args={"replay-user-messages": None}`.
- Rewinding after the stream ends requires resuming the session and calling `rewind_files()` on the live connection.

## Deployment and hosting rules

- Treat the SDK like a long-lived, stateful tool runner rather than a stateless completion API.
- For production, prefer sandboxed containers with controlled filesystem and network access.
- Recommend explicit `max_turns`, permission strategy, and monitoring instead of leaving long-running agents unconstrained.
- Choose between ephemeral, long-running, hybrid, or shared-container patterns based on the workload shape.

## Repo-aware guidance

If the current repository already uses the SDK, align with the local pattern before inventing a fresh abstraction.

In worker-oriented codebases, prefer this order of thought:
1. define prompt and output schema
2. wire `ClaudeAgentOptions`
3. attach MCP servers if needed
4. stream messages
5. pull the final `ResultMessage`
6. normalize `structured_output`
7. validate and persist the result

When the repo is cross-platform, keep subprocess command names portable. On Windows, external commands may need `.exe` while Unix-like systems usually do not.

## Response structure

For non-trivial requests, answer in this order:
1. **Recommended SDK surface** — `query()` or `ClaudeSDKClient`, with a short why
2. **Minimal working code** — small and correct before adding extras
3. **Permissions / MCP notes** — only the caveats that matter here
4. **Common pitfalls** — especially around tools, hooks, and output parsing
5. **Adaptation notes** — how to fit the code into the user's repo or worker

## What good help looks like

Good output from this skill should:
- choose the right SDK surface instead of mixing both patterns casually
- use exact Anthropic SDK names
- avoid hallucinating unsupported Python permission behavior
- explain when streaming input is required
- distinguish session continuity from filesystem rollback
- know when subagents are appropriate and what they inherit
- explain when to use in-process SDK tools versus external MCP servers
- show how structured output is actually retrieved
- stay grounded in current official docs and the user's existing codebase

## Examples and reusable patterns

See `references/patterns.md` for ready-to-adapt snippets covering:
- worker-style `query()` with structured output
- `ClaudeSDKClient` with a custom in-process tool
- a hook that blocks `.env` writes safely
- the permission behaviors most likely to trip people up

See `references/advanced-topics.md` when the task involves:
- sessions, streaming input, or `ClaudeSDKClient` lifecycle
- `AskUserQuestion`, approvals, and `can_use_tool`
- subagents, context inheritance, or delegation bugs
- agent loop behavior, result subtypes, cost/turn limits, or context growth

See `references/runtime-and-deployment.md` when the task involves:
- MCP transport choices, auth, tool search, or connection troubleshooting
- file checkpointing and rewind flows
- hosting, sandboxing, or production deployment patterns
