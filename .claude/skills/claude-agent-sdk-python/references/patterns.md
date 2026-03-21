# Claude Agent SDK Python Patterns

Use these patterns as starting points, then adapt them to the user's repo rather than pasting them blindly.

## 1) Worker-style `query()` with structured output

```python
import anyio
from pydantic import BaseModel
from claude_agent_sdk import ClaudeAgentOptions, ResultMessage, query


class CrawlResult(BaseModel):
    summary: str
    source_urls: list[str]


async def run_job() -> CrawlResult:
    options = ClaudeAgentOptions(
        system_prompt="You are a careful research agent.",
        output_format={
            "type": "json_schema",
            "schema": CrawlResult.model_json_schema(),
        },
        max_turns=10,
    )

    async for message in query(
        prompt="Research the official admissions page and summarize it.",
        options=options,
    ):
        if isinstance(message, ResultMessage):
            result = message.structured_output
            if isinstance(result, CrawlResult):
                return result
            if isinstance(result, dict):
                return CrawlResult.model_validate(result)
            raise TypeError(f"Unexpected structured output type: {type(result)!r}")

    raise RuntimeError("Agent returned no ResultMessage")


if __name__ == "__main__":
    anyio.run(run_job)
```

Why this pattern is good:

- it keeps the agent loop small
- it uses a schema instead of hand-wavy JSON instructions
- it handles the final `ResultMessage` explicitly

## 2) `ClaudeSDKClient` with an in-process SDK tool

```python
import asyncio
from typing import Any
from claude_agent_sdk import (
    ClaudeAgentOptions,
    ClaudeSDKClient,
    create_sdk_mcp_server,
    tool,
)


@tool("lookup_university", "Look up a university by name", {"name": str})
async def lookup_university(args: dict[str, Any]) -> dict[str, Any]:
    name = args["name"]
    return {
        "content": [
            {"type": "text", "text": f"Found internal record for {name}."}
        ]
    }


async def main() -> None:
    server = create_sdk_mcp_server(
        name="internal-tools",
        version="1.0.0",
        tools=[lookup_university],
    )

    options = ClaudeAgentOptions(
        mcp_servers={"internal": server},
        allowed_tools=["mcp__internal__lookup_university"],
        system_prompt="Use internal tools when they help.",
    )

    async with ClaudeSDKClient(options=options) as client:
        await client.query("Check whether UniMatch University exists in our records.")
        async for message in client.receive_response():
            print(message)


if __name__ == "__main__":
    asyncio.run(main())
```

Why this pattern is good:

- custom tool logic stays inside Python
- deployment is simpler than running a separate MCP subprocess
- the tool name is explicitly pre-approved

## 3) Hook that blocks writes to `.env`

```python
import asyncio
from claude_agent_sdk import ClaudeAgentOptions, ClaudeSDKClient, HookMatcher


async def protect_env_files(input_data, tool_use_id, context):
    if input_data["hook_event_name"] != "PreToolUse":
        return {}

    if input_data["tool_name"] not in {"Write", "Edit"}:
        return {}

    file_path = input_data["tool_input"].get("file_path", "")
    if file_path.replace("\\", "/").endswith("/.env") or file_path.endswith(".env"):
        return {
            "systemMessage": "Do not modify environment secret files.",
            "hookSpecificOutput": {
                "hookEventName": "PreToolUse",
                "permissionDecision": "deny",
                "permissionDecisionReason": "Editing .env files is blocked.",
            },
        }

    return {}


async def main() -> None:
    options = ClaudeAgentOptions(
        hooks={
            "PreToolUse": [
                HookMatcher(matcher="Write|Edit", hooks=[protect_env_files])
            ]
        }
    )

    async with ClaudeSDKClient(options=options) as client:
        await client.query("Write TEST=1 into .env")
        async for message in client.receive_response():
            print(message)


if __name__ == "__main__":
    asyncio.run(main())
```

Why this pattern is good:

- the matcher filters on tool names only
- file-path logic lives inside the callback where it belongs
- the deny response includes `hookEventName` and a reason

## 4) Permission notes worth remembering

- `tools` controls the base built-in tool set; `allowed_tools` only pre-approves from that available surface.
- `allowed_tools` pre-approves tools; it does not remove other tools from Claude's available toolset.
- `disallowed_tools` is the reliable way to block tools.
- Python does not support `dontAsk` permission mode.
- `bypassPermissions` is broad; do not pair it with weak assumptions about `allowed_tools`.
- Use `setting_sources=["project"]` when SDK code should load `.claude/settings.json` permissions or hooks.

## 5) `can_use_tool` for runtime permission control

```python
import asyncio
from collections.abc import AsyncIterator

from claude_agent_sdk import (
    ClaudeAgentOptions,
    HookMatcher,
    PermissionResultAllow,
    PermissionResultDeny,
    ToolPermissionContext,
    query,
)


async def prompt_stream() -> AsyncIterator[dict]:
    yield {
        "type": "user",
        "message": {
            "role": "user",
            "content": "Create hello.py, but never modify .env files.",
        },
        "parent_tool_use_id": None,
        "session_id": "permission-demo",
    }


async def permission_callback(
    tool_name: str,
    input_data: dict,
    context: ToolPermissionContext,
) -> PermissionResultAllow | PermissionResultDeny:
    if tool_name in {"Read", "Glob", "Grep"}:
        return PermissionResultAllow()

    if tool_name in {"Write", "Edit", "MultiEdit"}:
        file_path = input_data.get("file_path", "")
        if file_path.replace("\\", "/").endswith("/.env") or file_path.endswith(".env"):
            return PermissionResultDeny(message="Editing .env files is blocked.")

    return PermissionResultAllow()


async def keep_stream_open(input_data, tool_use_id, context):
    return {"continue_": True}


async def main() -> None:
    async for message in query(
        prompt=prompt_stream(),
        options=ClaudeAgentOptions(
            can_use_tool=permission_callback,
            permission_mode="default",
            hooks={"PreToolUse": [HookMatcher(matcher=None, hooks=[keep_stream_open])]},
            cwd=".",
        ),
    ):
        print(message)


if __name__ == "__main__":
    asyncio.run(main())
```

Why this pattern is good:

- `can_use_tool` is used for runtime approval instead of overloading hooks for everything
- the example is explicit about streaming mode, which `can_use_tool` requires
- it includes the documented Python workaround hook to keep the stream open for approval callbacks
- deny logic stays narrow and easy to audit
