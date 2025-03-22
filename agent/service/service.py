import logging
import asyncio
import threading
from threading import Lock
from typing import Dict, Any
import time

from agents import Runner, RunResult
from agent.clients import ConfigurationClient
from agent.models import create_agent_for_user
from agent.protos import agent_pb2, agent_pb2_grpc

logger = logging.getLogger(__name__)


class AgentServicer(agent_pb2_grpc.AgentServiceServicer):
    def __init__(self, configuration_service: ConfigurationClient, cleanup_interval: int = 3600):
        self.configuration_service = configuration_service
        self.user_conversations: Dict[str, Dict[str, Any]] = {}
        self.lock = Lock()
        self.cleanup_interval = cleanup_interval

        self.loop = asyncio.new_event_loop()
        self.loop_thread = threading.Thread(target=self._run_loop, daemon=True)
        self.loop_thread.start()

        self.cleanup_thread = threading.Thread(target=self._cleanup_inactive_conversations, daemon=True)
        self.cleanup_thread.start()

    def _run_loop(self):
        asyncio.set_event_loop(self.loop)
        self.loop.run_forever()

    def _cleanup_inactive_conversations(self):
        while True:
            time.sleep(self.cleanup_interval)
            current_time = time.time()
            to_remove = []

            with self.lock:
                for user_id, data in self.user_conversations.items():
                    if current_time - data.get("last_activity", 0) > self.cleanup_interval:
                        to_remove.append(user_id)

                for user_id in to_remove:
                    del self.user_conversations[user_id]

    async def _process_message(self, user_id: str, message: str) -> str:
        with self.lock:
            conversation = self.user_conversations.get(user_id)
            if not conversation:
                raise ValueError(f"No active conversation found for user {user_id}")

            conversation["last_activity"] = time.time()
            agent = conversation["agent"]
            history = conversation["history"]

        result: RunResult = await Runner.run(agent, input=message, context=history)

        with self.lock:
            if user_id in self.user_conversations:
                self.user_conversations[user_id]["history"].append(result.final_output)
                self.user_conversations[user_id]["last_activity"] = time.time()

        return result.final_output

    async def _initialize_agent(self, user_id: str) -> str:
        try:
            if not user_id or not isinstance(user_id, str) or not user_id.strip():
                return "ERROR: Invalid user ID"

            config = self.configuration_service.get_configuration(user_id)
            if not config:
                return f"ERROR: Configuration not found for user {user_id}"

            agent = create_agent_for_user(config)

            with self.lock:
                self.user_conversations[user_id] = {
                    "agent": agent,
                    "history": [],
                    "last_activity": time.time()
                }
            return "Agent conversation initialized successfully"
        except Exception as e:
            logger.exception(f"Failed to create agent for user {user_id}")
            return f"ERROR: Failed to create agent: {str(e)}"

    async def _close_agent(self, user_id: str) -> None:
        with self.lock:
            if user_id in self.user_conversations:
                try:
                    agent = self.user_conversations[user_id].get("agent")
                    if hasattr(agent, "cleanup") and callable(agent.cleanup):
                        await agent.cleanup()
                except Exception as e:
                    logger.error(f"Error during agent cleanup for user {user_id}: {e}")

                del self.user_conversations[user_id]

    def AgentWebsocketStream(self, request_iterator, context):
        user_id = None

        for request in request_iterator:
            try:
                if not hasattr(request, "type"):
                    yield agent_pb2.AgentMessage(
                        type=agent_pb2.MessageType.ERROR,
                        content="Invalid request format: missing type",
                    )
                    continue

                if not user_id and request.type != agent_pb2.MessageType.INITIALIZE:
                    yield agent_pb2.AgentMessage(
                        type=agent_pb2.MessageType.ERROR,
                        content="Conversation must be initialized first",
                    )
                    continue

                if request.type == agent_pb2.MessageType.INITIALIZE:
                    user_id = request.user_id

                    result_future = asyncio.run_coroutine_threadsafe(
                        self._initialize_agent(user_id), self.loop
                    )
                    result = result_future.result()

                    message_type = (
                        agent_pb2.MessageType.ERROR
                        if result.startswith("ERROR:")
                        else agent_pb2.MessageType.AGENT_RESPONSE
                    )
                    yield agent_pb2.AgentMessage(
                        type=message_type,
                        user_id=user_id,
                        content=result,
                        metadata=request.metadata,
                    )

                elif request.type == agent_pb2.MessageType.USER_MESSAGE:
                    if not request.content or not request.content.strip():
                        yield agent_pb2.AgentMessage(
                            type=agent_pb2.MessageType.ERROR,
                            user_id=user_id,
                            content="Empty message received",
                            metadata=request.metadata,
                        )
                        continue

                    try:
                        response_future = asyncio.run_coroutine_threadsafe(
                            self._process_message(user_id, request.content), self.loop
                        )
                        response_text = response_future.result()
                        yield agent_pb2.AgentMessage(
                            type=agent_pb2.MessageType.AGENT_RESPONSE,
                            user_id=user_id,
                            content=response_text,
                            metadata=request.metadata,
                        )
                    except ValueError as e:
                        yield agent_pb2.AgentMessage(
                            type=agent_pb2.MessageType.ERROR,
                            user_id=user_id,
                            content=str(e),
                            metadata=request.metadata,
                        )
                    except Exception as e:
                        logger.exception(f"Error processing message for user {user_id}")
                        yield agent_pb2.AgentMessage(
                            type=agent_pb2.MessageType.ERROR,
                            user_id=user_id,
                            content=f"Error processing message: {str(e)}",
                            metadata=request.metadata,
                        )

                elif request.type == agent_pb2.MessageType.CLOSE:
                    close_future = asyncio.run_coroutine_threadsafe(
                        self._close_agent(user_id), self.loop
                    )
                    close_future.result()
                    yield agent_pb2.AgentMessage(
                        type=agent_pb2.MessageType.AGENT_RESPONSE,
                        user_id=user_id,
                        content="Conversation closed successfully",
                        metadata=request.metadata,
                    )
                    break

                else:
                    yield agent_pb2.AgentMessage(
                        type=agent_pb2.MessageType.ERROR,
                        user_id=user_id or "",
                        content="Unknown request type",
                        metadata=request.metadata,
                    )

            except Exception as e:
                yield agent_pb2.AgentMessage(
                    type=agent_pb2.MessageType.ERROR,
                    user_id=user_id or "",
                    content=f"Internal server error: {str(e)}",
                )

        logger.info(f"Websocket stream complete for user {user_id}")

        if user_id:
            try:
                close_future = asyncio.run_coroutine_threadsafe(
                    self._close_agent(user_id), self.loop
                )
                close_future.result()
            except Exception as e:
                logger.error(f"Error auto-closing conversation for user {user_id}: {e}")

    def __del__(self):
        try:
            self.loop.call_soon_threadsafe(self.loop.stop)
            self.loop_thread.join(timeout=5)

            for user_id in list(self.user_conversations.keys()):
                close_future = asyncio.run_coroutine_threadsafe(
                    self._close_agent(user_id), self.loop
                )
                close_future.result(timeout=5)
        except Exception as e:
            logger.error(f"Error during AgentServicer cleanup: {e}")