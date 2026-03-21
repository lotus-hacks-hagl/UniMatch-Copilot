# Claude Agent SDK Python Advanced Topics

Use this file when the user's question goes beyond a simple one-shot `query()` example.

## Sessions and input modes

### Choose the right interaction style

- Use plain `query(prompt="...")` for one-shot or worker-style jobs.
- Use streaming input when the app needs follow-up messages, interruptions, image attachments, or approval handling.
- Use `ClaudeSDKClient` in Python when one process should keep a session alive across multiple prompts naturally.

### Session rules worth remembering

- `ClaudeSDKClient` automatically continues the same session across multiple `client.query(...)` calls.
- With `query()`, capture `ResultMessage.session_id` if you will need `resume` later.
- `fork_session=True` branches conversation history, not filesystem state.
- If `resume` restores the wrong context or a fresh session, check `cwd` first.
- Python persists sessions to disk; there is no stateless `persistSession: false` equivalent.

## Approvals and AskUserQuestion

### What `can_use_tool` actually does

`can_use_tool` is the Python runtime approval callback for both:

- tool approval requests
- clarifying questions via `AskUserQuestion`

### Python-specific caveats

- `can_use_tool` requires streaming mode.
- In Python, the official docs currently require a dummy `PreToolUse` hook that returns `{"continue_": True}` so the stream stays open long enough for the callback.
- If you restrict the built-in `tools` set and still want clarifying questions, include `AskUserQuestion` in that list.

### Mapping answers for clarifying questions

When handling `AskUserQuestion`, return:

- the original `questions` array
- an `answers` object mapping each `question` string to the chosen label text

For multi-select questions, join labels with `", "`.

## Subagents

### When to use them

Use subagents when the user needs:

- context isolation
- parallelized investigation
- specialized prompts or tool restrictions

### Rules that matter

- The parent needs `Agent` in `allowed_tools` for delegation to happen.
- `AgentDefinition.description` is routing metadata; write it as “when to use this agent”.
- Subagents do not inherit the parent conversation history.
- Subagents cannot spawn their own subagents; do not give them the `Agent` tool.
- Programmatic agents are the recommended SDK path.

### Practical debugging tips

If Claude refuses to delegate:

- verify `Agent` is allowed
- write a clearer description
- explicitly ask for the subagent by name in the prompt

Windows note:

- very long inline subagent prompts can fail because of command-line length limits

## Structured outputs

### Good practice

- Prefer Pydantic models and `.model_json_schema()` in Python.
- Use focused schemas with realistic required fields.
- Validate returned dicts with `model_validate()` when you want typed objects in app code.

### Failure mode to handle

When the schema is too hard to satisfy or the task is ambiguous, check for:

- `ResultMessage.subtype == "error_max_structured_output_retries"`

Suggested recovery strategies:

- simplify the schema
- make uncertain fields optional
- tighten the prompt
- fall back to unstructured output if needed

## Agent loop and result handling

### Core lifecycle

The SDK yields a stream that generally includes:

- `SystemMessage` init metadata
- `AssistantMessage` turns
- `UserMessage` tool results and streamed user inputs
- optional `StreamEvent` partial events
- final `ResultMessage`

### Result subtypes that matter in app code

- `success`
- `error_max_turns`
- `error_max_budget_usd`
- `error_during_execution`
- `error_max_structured_output_retries`

Always inspect `subtype` before assuming `result` exists.

### Efficiency guidance

- Set `max_turns` for open-ended tasks.
- Set `max_budget_usd` for production agents.
- Use lower `effort` for routine tasks and higher `effort` for debugging/refactors.
- Use subagents and MCP tool search to reduce context bloat.
