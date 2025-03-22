import logging
import asyncio
import threading
from threading import Lock
from typing import Dict, Any

from agents import Runner, RunResult
from agent.clients import ConfigurationClient
from agent.models import create_agent_for_user
from agent.protos import agent_pb2, agent_pb2_grpc

logger = logging.getLogger(__name__)

class AgentServicer(agent_pb2_grpc.AgentServiceServicer):
    def __init__(self, configuration_service: ConfigurationClient):
        self.configuration_service = configuration_service
        self.user_conversations: Dict[str, Dict[str, Any]] = {}
        self.lock = Lock()

        self.loop = asyncio.new_event_loop()

        self.loop_thread = threading.Thread(target=self._run_event_loop, daemon=True)
        self.loop_thread.start()
        logger.info("Event loop started in background thread")

    def _run_event_loop(self):
        asyncio.set_event_loop(self.loop)
        self.loop.run_forever()

    def __del__(self):
        logger.info("Cleaning up AgentServicer resources")
        if hasattr(self, 'loop') and self.loop.is_running():
            self.loop.call_soon_threadsafe(self.loop.stop)
            if hasattr(self, 'loop_thread'):
                self.loop_thread.join(timeout=5)

    async def _process_message(self, user_id: str, message: str) -> str:
        conversation_data = self.user_conversations.get(user_id)
        if not conversation_data:
            raise ValueError(f"No active conversation found for user {user_id}")

        agent = conversation_data["agent"]
        history = conversation_data["history"]

        result: RunResult = await Runner.run(
            agent,
            input=message,
            context=history
        )

        conversation_data["history"].append(result.final_output)
        return result.final_output

    async def _initialize_agent(self, user_id: str) -> str:
        try:
            config = self.configuration_service.get_configuration(user_id)
            if not config:
                return f"ERROR: Configuration not found for user {user_id}"

            agent = create_agent_for_user(config)

            with self.lock:
                self.user_conversations[user_id] = {
                    "agent": agent,
                    "history": []
                }

            logger.info(f"Created new conversation for user {user_id}")
            return "Agent conversation initialized successfully"

        except Exception as e:
            logger.error(f"Failed to create agent for user {user_id}: {str(e)}")
            return f"ERROR: Failed to create agent: {str(e)}"

    async def _close_agent(self, user_id: str) -> None:
        with self.lock:
            if user_id in self.user_conversations:
                del self.user_conversations[user_id]
                logger.info(f"Closed conversation for user {user_id}")

    async def _handle_stream(self, request_iterator, response_queue):
        user_id = None
        try:
            for request in request_iterator:
                logger.info(f"Processing request: {request}")

                if not user_id and request.type != agent_pb2.MessageType.INITIALIZE:
                    await response_queue.put(agent_pb2.AgentMessage(
                        type=agent_pb2.MessageType.ERROR,
                        content="Conversation must be initialized first"
                    ))
                    continue

                if request.type == agent_pb2.MessageType.INITIALIZE:
                    user_id = request.user_id
                    logger.info(f"Initializing agent for user: {user_id}")
                    result = await self._initialize_agent(user_id)

                    if result.startswith("ERROR:"):
                        await response_queue.put(agent_pb2.AgentMessage(
                            type=agent_pb2.MessageType.ERROR,
                            user_id=user_id,
                            content=result
                        ))
                    else:
                        await response_queue.put(agent_pb2.AgentMessage(
                            type=agent_pb2.MessageType.AGENT_RESPONSE,
                            user_id=user_id,
                            content=result
                        ))

                elif request.type == agent_pb2.MessageType.USER_MESSAGE:
                    try:
                        logger.info(f"Processing user message from {user_id}: {request.content[:50]}...")
                        user_message_response = await self._process_message(user_id, request.content)
                        await response_queue.put(agent_pb2.AgentMessage(
                            type=agent_pb2.MessageType.AGENT_RESPONSE,
                            user_id=user_id,
                            content=user_message_response,
                            metadata=request.metadata
                        ))
                    except Exception as e:
                        logger.error(f"Error processing message: {str(e)}")
                        await response_queue.put(agent_pb2.AgentMessage(
                            type=agent_pb2.MessageType.ERROR,
                            user_id=user_id,
                            content=f"Error processing message: {str(e)}"
                        ))

                elif request.type == agent_pb2.MessageType.CLOSE:
                    logger.info(f"Closing conversation for user {user_id}")
                    await self._close_agent(user_id)
                    await response_queue.put(agent_pb2.AgentMessage(
                        type=agent_pb2.MessageType.AGENT_RESPONSE,
                        user_id=user_id,
                        content="Conversation closed successfully"
                    ))
                    break

        except Exception as e:
            logger.error(f"Stream error: {str(e)}")
            if user_id:
                await response_queue.put(agent_pb2.AgentMessage(
                    type=agent_pb2.MessageType.ERROR,
                    user_id=user_id,
                    content=f"Stream error: {str(e)}"
                ))
                await self._close_agent(user_id)

        finally:
            logger.info("Stream processing complete, sending None to terminate")
            if user_id and user_id in self.user_conversations:
                await self._close_agent(user_id)
            await response_queue.put(None)

    def AgentWebsocketStream(self, request_iterator, context):
        logger.info("New websocket stream connection established")
        response_queue = asyncio.Queue()

        future = asyncio.run_coroutine_threadsafe(
            self._handle_stream(request_iterator, response_queue),
            self.loop
        )

        def on_done(fut):
            try:
                fut.result()
            except Exception as e:
                logger.error(f"Unhandled exception in handle_stream: {e}")

        future.add_done_callback(on_done)

        try:
            while True:
                get_future = asyncio.run_coroutine_threadsafe(response_queue.get(), self.loop)
                response = get_future.result()

                if response is None:
                    logger.info("Received None response, ending stream")
                    break

                logger.info(f"Yielding response: type={response.type}")
                yield response

        except Exception as e:
            logger.error(f"Error in response generator: {e}")

        logger.info("Websocket stream complete")
