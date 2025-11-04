import logging
import json
from typing import Dict, Any, List, Optional, Callable, Awaitable
from datetime import datetime
from langchain_openai import ChatOpenAI
from langchain_core.messages import SystemMessage, HumanMessage, AIMessage
from langfuse.langchain import CallbackHandler

logger = logging.getLogger(__name__)


class CoverLetterWriter:
    """Specialized agent for writing cover letters"""

    def __init__(self, resume_context: str, resume_name: str):
        self.resume_context = resume_context
        self.resume_name = resume_name
        self.llm = None
        self.send_ws: Optional[Callable[[str], Awaitable[None]]] = None

    async def setup(self):
        """Initialize the LLM for cover letter writing"""
        logger.info("Setting up Cover Letter Writer LLM...")
        langfuse_handler = CallbackHandler()

        self.llm = ChatOpenAI(
            model="gpt-5-chat-latest",
            temperature=0.7,
            max_completion_tokens=1500,
            max_retries=3,
            callbacks=[langfuse_handler],
        )
        logger.info("✓ Cover Letter Writer LLM configured")

    def set_websocket_sender(self, send_ws: Callable[[str], Awaitable[None]]):
        """Set the WebSocket sender for streaming responses"""
        self.send_ws = send_ws

    def _get_system_prompt(self) -> str:
        """Generate system prompt for cover letter writing"""
        return f"""You are a professional cover letter writer for {self.resume_name}.

AVAILABLE CONTEXT:
{self.resume_context}

YOUR TASK:
Write compelling, professional cover letters in Russian that:
1. Are properly formatted (greeting, 2-4 body paragraphs, closing)
2. Reference specific job details from the posting. Be sure to ask for the job description.
3. Use actual experience and skills from the resume context above
4. Match the professional tone from sample cover letters
5. Are 250-400 words in length
6. Use plain text formatting (no markdown)
7. Are tailored to the specific job and company
8. Only output the contents of the cover letter
9. IMPORTANT! Respond in Russian

STYLE GUIDELINES:
- Professional but personable tone
- Specific examples over generic claims
- Show enthusiasm for the role
- Connect resume experience to job requirements
- Strong opening and closing

Date: {datetime.now().strftime("%Y-%m-%d")}
"""

    async def write_cover_letter(
        self, job_details: str, thread_id: str = None, conversation_history: List = None
    ) -> Dict[str, Any]:
        """Write a cover letter based on job details and conversation history"""
        logger.info(f"📝 Cover Letter Writer: Starting cover letter generation")
        logger.info(f"   Job details: {job_details[:100]}...")

        system_message = SystemMessage(content=self._get_system_prompt())

        # Build context from conversation history if available
        if conversation_history and len(conversation_history) > 1:
            # Extract relevant context from previous messages
            context_parts = []
            for msg in conversation_history[:-1]:  # Exclude the current message
                if isinstance(msg, HumanMessage):
                    context_parts.append(f"User: {msg.content}")
                elif isinstance(msg, AIMessage):
                    context_parts.append(f"Assistant: {msg.content}")

            conversation_context = "\n".join(
                context_parts[-4:]
            )  # Last 4 messages for context
            user_message = HumanMessage(
                content=f"Previous conversation:\n{conversation_context}\n\nNow write a cover letter based on the job information provided above."
            )
        else:
            user_message = HumanMessage(
                content=f"Write a cover letter for the following job:\n\n{job_details}"
            )

        messages = [system_message, user_message]

        # If no websocket available, just invoke directly
        if not self.send_ws:
            logger.info("   Generating cover letter (non-streaming)")
            response = await self.llm.ainvoke(messages)
            return {"content": response.content}

        # Streaming path
        logger.info("   Generating cover letter (streaming)")
        await self.send_ws(json.dumps({"type": "start"}))

        full_content = ""
        async for chunk in self.llm.astream(messages):
            if hasattr(chunk, "content") and chunk.content:
                full_content += chunk.content
                await self.send_ws(
                    json.dumps(
                        {"type": "content", "content": full_content, "is_final": False}
                    )
                )

        await self.send_ws(
            json.dumps({"type": "content", "content": full_content, "is_final": True})
        )
        await self.send_ws(json.dumps({"type": "end"}))

        logger.info(f"✅ Cover letter generated: {len(full_content)} characters")

        return {"content": full_content}
