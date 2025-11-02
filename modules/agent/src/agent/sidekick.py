from langgraph.checkpoint.memory import MemorySaver
import uuid
import json
import logging
from typing import List, Any, Optional, Dict, Annotated, Callable, Awaitable
from typing_extensions import TypedDict
from agent.sidekick_tools import all_tools
from langgraph.graph import StateGraph, START, END
from langchain_openai import ChatOpenAI
from pydantic import BaseModel, Field
from dotenv import load_dotenv
from langgraph.graph.message import add_messages
from langgraph.prebuilt import ToolNode
from datetime import datetime
from langchain_core.messages import AIMessage, HumanMessage, SystemMessage, ToolMessage

logger = logging.getLogger(__name__)
load_dotenv(override=True)


class State(TypedDict):
    messages: Annotated[List[Any], add_messages]
    success_criteria: str
    feedback_on_work: Optional[str]
    success_criteria_met: bool
    user_input_needed: bool
    resume_name: str
    resume_context: str  # Added to hold resume, summary, and cover letters


class EvaluatorOutput(BaseModel):
    feedback: str = Field(description="Feedback on the assistant's response")
    success_criteria_met: bool = Field(
        description="Whether the success criteria have been met"
    )
    user_input_needed: bool = Field(
        description="True if more input is needed from the user, or clarifications, or the assistant is stuck"
    )


class Sidekick:
    def __init__(
        self, resume_text: str = "", summary: str = "", cover_letters: list = None
    ):
        self.worker_llm_with_tools = None
        self.evaluator_llm_with_output = None
        self.tools = None
        self.graph = None
        self.sidekick_id = str(uuid.uuid4())
        self.memory = MemorySaver()
        self.send_ws: Optional[Callable[[str], Awaitable[None]]] = None

        # Store document context
        self.resume_text = resume_text
        self.summary = summary
        self.cover_letters = cover_letters or []

        # Build resume context string
        self.resume_context = self._build_resume_context()

    def _build_resume_context(self) -> str:
        """Build the resume context string from loaded documents"""
        context_parts = []

        if self.summary:
            context_parts.append(f"SUMMARY\n{self.summary}")

        if self.resume_text:
            context_parts.append(f"RESUME\n{self.resume_text}")

        if self.cover_letters:
            cover_letters_text = "\n\n".join(self.cover_letters)
            context_parts.append(
                f"COVER_LETTERS (for reference on writing style)\n{cover_letters_text}"
            )

        return "\n\n".join(context_parts)

    async def setup(self):
        """Initialize the LLMs and build the graph"""
        self.tools = await all_tools()
        worker_llm = ChatOpenAI(model="gpt-5-chat-latest", temperature=0)
        self.worker_llm_with_tools = worker_llm.bind_tools(self.tools)

        evaluator_llm = ChatOpenAI(model="gpt-5-chat-latest", temperature=0)
        self.evaluator_llm_with_output = evaluator_llm.with_structured_output(
            EvaluatorOutput
        )
        await self.build_graph()
        logger.info(f"Sidekick initialized with {len(self.tools)} tools")

    def set_websocket_sender(self, send_ws: Callable[[str], Awaitable[None]]):
        """Set the WebSocket sender for streaming responses"""
        self.send_ws = send_ws

    async def _stream_if_available(self, content: str, is_final: bool = False):
        """Stream content to WebSocket if available"""
        if self.send_ws:
            payload = {"type": "content", "content": content, "is_final": is_final}
            await self.send_ws(json.dumps(payload))

    def worker(self, state: State) -> Dict[str, Any]:
        """The worker node that interacts with the LLM and tools"""
        system_message = f"""
ROLE
You are {state['resume_name']} on his personal website. Answer questions about his career, background, skills, and experience. Speak as {state['resume_name']}, professionally and succinctly.

You have access to detailed information about {state['resume_name']}'s background below. Use this information to answer questions accurately.

DATA ACCESS
{state['resume_context']}

CURRENT TASK
{state['success_criteria']}

STYLE (MUST FOLLOW)
- Output must be plain text only (no Markdown or other formatting).
- Do not use asterisks, underscores, backticks, tildes, hashes, brackets, angle brackets, emojis, or decorative characters.
- Do not bold, italicize, add headings, lists, tables, code fences, links, or inline formatting.
- Keep responses concise; short paragraphs separated by newlines only.
- Do not re-greet mid-conversation. Do not say “Hi there!” again after the first turn.

INSTRUCTIONS
- Use the resume data above to answer questions
- Be conversational but professional
- Keep responses focused and concise
- If the user provides contact information, use the record_user_details tool
- If you cannot answer a question with the available information, use record_unknown_question tool

JOB SEARCH
- When prompted to job search, search for roles matching your resume and experience. Focus on full-stack engineering roles or AI roles.
- If you need to search for current information (like job postings), use the search tool to return direct URL links to specific job postings (no job board queries).
- Prioritize recent job postings (within the last 7 days)
- Search for roles in New York City.
- Prompt the user if you wish to expand the search to different criteria, but always remain within New York City.
- If the user asks you to conduct a job search, do not prompt them for their email or contact information.

CONTACT CAPTURE (ASK ONCE, RECORD ONCE) 
- After your third substantive answer, briefly ask once for the user’s name and email for follow-up. 
- If an email is provided (with or without a name), immediately call record_user_details with provided fields. 
- If a name is provided but no email, ask once for the email. If declined, acknowledge and continue. Do not ask again once asked or once recorded. 

UNKNOWN ANSWERS 
- If you do not know an answer, call record_unknown_question with the question, then (only if not already collected) ask once for name/email as above. 

SCOPE CONTROL 
- If a request is not about {state['resume_name']}'s career, background, skills, or experience, briefly steer back to those topics. 

COVER LETTER WORKFLOW 
- If asked to write a cover letter, ask for a job-posting URL, read it, then write the letter directly using COVER_LETTERS for style. Do not ask for confirmation.

The current date and time is {datetime.now().strftime("%Y-%m-%d %H:%M:%S")}
"""

        if state.get("feedback_on_work"):
            system_message += f"""

FEEDBACK
Previously you thought you answered the question, but your reply was rejected because the success criteria was not met.
Here is the feedback on why this was rejected:
{state['feedback_on_work']}

With this feedback, please continue to try to answer the question, ensuring that you meet the success criteria or have a question for the user.
"""

        # Update or add system message
        messages = state["messages"].copy()
        found_system_message = False
        for i, message in enumerate(messages):
            if isinstance(message, SystemMessage):
                messages[i] = SystemMessage(content=system_message)
                found_system_message = True
                break

        if not found_system_message:
            messages = [SystemMessage(content=system_message)] + messages

        # Invoke the LLM with tools
        response = self.worker_llm_with_tools.invoke(messages)

        return {"messages": [response]}

    def worker_router(self, state: State) -> str:
        """Route based on whether tools need to be called"""
        last_message = state["messages"][-1]

        if hasattr(last_message, "tool_calls") and last_message.tool_calls:
            return "tools"
        return "evaluator"

    def format_conversation(self, messages: List[Any]) -> str:
        """Format conversation history for the evaluator"""
        conversation = "Conversation history:\n\n"
        for message in messages:
            if isinstance(message, HumanMessage):
                conversation += f"User: {message.content}\n"
            elif isinstance(message, AIMessage):
                text = message.content or "[Tool use]"
                conversation += f"Assistant: {text}\n"
            elif isinstance(message, ToolMessage):
                conversation += f"[Tool result: {message.content[:100]}...]\n"
        return conversation

    def evaluator(self, state: State) -> Dict[str, Any]:
        """Evaluate if the task is complete"""
        last_response = state["messages"][-1].content

        system_message = """You are an evaluator that determines if a task has been completed successfully by an Assistant.
Assess the Assistant's last response based on the given criteria. Respond with your feedback, and with your decision on whether the success criteria has been met,
and whether more input is needed from the user."""

        user_message = f"""You are evaluating a conversation between the User and Assistant. You decide what action to take based on the last response from the Assistant.

The entire conversation with the assistant, with the user's original request and all replies, is:
{self.format_conversation(state['messages'])}

The success criteria for this assignment is:
{state['success_criteria']}

And the final response from the Assistant that you are evaluating is:
{last_response}

Respond with your feedback, and decide if the success criteria is met by this response.
Also, decide if more user input is required, either because the assistant has a question, needs clarification, or seems to be stuck and unable to answer without help.

The Assistant has access to tools to record user details and unknown questions. Give the Assistant the benefit of the doubt if they say they've done something.
"""

        if state.get("feedback_on_work"):
            user_message += f"\nNote: In a prior attempt, you provided this feedback: {state['feedback_on_work']}\n"
            user_message += "If you're seeing the Assistant repeating the same mistakes, consider responding that user input is required."

        evaluator_messages = [
            SystemMessage(content=system_message),
            HumanMessage(content=user_message),
        ]

        eval_result = self.evaluator_llm_with_output.invoke(evaluator_messages)

        return {
            "feedback_on_work": eval_result.feedback,
            "success_criteria_met": eval_result.success_criteria_met,
            "user_input_needed": eval_result.user_input_needed,
        }

    def route_based_on_evaluation(self, state: State) -> str:
        """Route based on evaluation results"""
        if state["success_criteria_met"] or state["user_input_needed"]:
            return "END"
        return "worker"

    async def build_graph(self):
        """Build the LangGraph workflow"""
        graph_builder = StateGraph(State)

        # Add nodes
        graph_builder.add_node("worker", self.worker)
        graph_builder.add_node("tools", ToolNode(tools=self.tools))
        graph_builder.add_node("evaluator", self.evaluator)

        # Add edges
        graph_builder.add_conditional_edges(
            "worker", self.worker_router, {"tools": "tools", "evaluator": "evaluator"}
        )
        graph_builder.add_edge("tools", "worker")
        graph_builder.add_conditional_edges(
            "evaluator",
            self.route_based_on_evaluation,
            {"worker": "worker", "END": END},
        )
        graph_builder.add_edge(START, "worker")

        # Compile the graph
        self.graph = graph_builder.compile(checkpointer=self.memory)
        logger.info("Graph compiled successfully")

    async def run_streaming(
        self, message: str, thread_id: str, success_criteria: str = None
    ) -> Dict[str, Any]:
        """Run the graph with streaming support for WebSocket"""
        config = {"configurable": {"thread_id": thread_id}}

        state = {
            "messages": [HumanMessage(content=message)],
            "success_criteria": success_criteria
            or "Provide a clear and accurate response to the user's question",
            "feedback_on_work": None,
            "success_criteria_met": False,
            "user_input_needed": False,
            "resume_name": "William Hubenschmidt",
            "resume_context": self.resume_context,
        }

        # Stream events if WebSocket is available
        if self.send_ws:
            await self.send_ws(json.dumps({"type": "start"}))

            last_content = ""

            async for event in self.graph.astream(
                state, config=config, stream_mode="values"
            ):
                # Get the last message
                if "messages" in event and event["messages"]:
                    last_msg = event["messages"][-1]

                    # Stream AI responses
                    if isinstance(last_msg, AIMessage) and last_msg.content:
                        current_content = last_msg.content

                        # Only send if content changed
                        if current_content != last_content:
                            last_content = current_content
                            await self.send_ws(
                                json.dumps(
                                    {
                                        "type": "content",
                                        "content": current_content,
                                        "is_final": False,
                                    }
                                )
                            )

            # Send final message
            await self.send_ws(
                json.dumps(
                    {"type": "content", "content": last_content, "is_final": True}
                )
            )
            await self.send_ws(json.dumps({"type": "end"}))
        else:
            # No streaming, just invoke
            result = await self.graph.ainvoke(state, config=config)
            return result

        return state

    def free_resources(self):
        """Clean up resources"""
        logger.info("Freeing Sidekick resources")
        self.send_ws = None
