"""Logging hooks for agent execution observability."""

import logging
from typing import Any

from agents import RunHooks, Agent, Tool
from agents.run_context import RunContextWrapper

logger = logging.getLogger("agent.hooks")


class LoggingHooks(RunHooks):
    """Hooks that log agent execution steps to stdout."""

    async def on_agent_start(self, context: RunContextWrapper, agent: Agent) -> None:
        """Log when an agent starts execution."""
        logger.info(f"▶️  AGENT START: {agent.name}")

    async def on_agent_end(self, context: RunContextWrapper, agent: Agent, output: Any) -> None:
        """Log when an agent completes with output."""
        output_preview = str(output)[:100] + "..." if len(str(output)) > 100 else str(output)
        logger.info(f"✅ AGENT END: {agent.name} → {output_preview}")

        # Log token usage if available
        try:
            usage = context.usage
            if usage:
                logger.info(
                    f"📊 TOKENS [{agent.name}]: in={usage.input_tokens} "
                    f"out={usage.output_tokens} total={usage.total_tokens}"
                )
        except (AttributeError, TypeError):
            pass

    async def on_handoff(
        self, context: RunContextWrapper, from_agent: Agent, to_agent: Agent
    ) -> None:
        """Log when control transfers between agents."""
        logger.info(f"🔀 HANDOFF: {from_agent.name} → {to_agent.name}")

    async def on_tool_start(
        self, context: RunContextWrapper, agent: Agent, tool: Tool
    ) -> None:
        """Log when a tool is about to be called."""
        logger.info(f"🔧 TOOL START: {tool.name}")

    async def on_tool_end(
        self, context: RunContextWrapper, agent: Agent, tool: Tool, result: str
    ) -> None:
        """Log when a tool completes."""
        result_preview = result[:80] + "..." if len(result) > 80 else result
        logger.info(f"🔧 TOOL END: {tool.name} → {result_preview}")

    async def on_llm_start(
        self, context: RunContextWrapper, agent: Agent, *args, **kwargs
    ) -> None:
        """Log when LLM call begins."""
        logger.debug(f"💭 LLM START: {agent.name}")

    async def on_llm_end(
        self, context: RunContextWrapper, agent: Agent, *args, **kwargs
    ) -> None:
        """Log when LLM call completes."""
        logger.debug(f"💭 LLM END: {agent.name}")


# Singleton instance for reuse
logging_hooks = LoggingHooks()
