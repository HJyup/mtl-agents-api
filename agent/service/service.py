import logging
import asyncio
import grpc
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

    def CreateAgentStream(self, request, context):
        user_id = request.user_id

        try:
            config = self.configuration_service.get_configuration(user_id)
            if not config:
                context.set_code(grpc.StatusCode.NOT_FOUND)
                context.set_details(f"Configuration not found for user {user_id}")
                return

            agent = create_agent_for_user(config)
        except Exception as e:
            logger.error(f"Failed to create agent for user {user_id}: {str(e)}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Failed to create agent: {str(e)}")
            return

        with self.lock:
            self.user_conversations[user_id] = {
                "agent": agent,
                "history": []
            }

        logger.info(f"Created new conversation for user {user_id}")

        yield agent_pb2.AgentStreamResponse(
            message="Agent conversation initialized"
        )

    def SendAgentMessage(self, request, context):
        user_id = request.user_id
        message = request.message

        logger.info(f"Received message from user {user_id}")

        with self.lock:
            if user_id not in self.user_conversations:
                context.set_code(grpc.StatusCode.NOT_FOUND)
                context.set_details(
                    f"No active conversation for user {user_id}"
                )
                return agent_pb2.SendAgentMessageResponse()

        try:
            loop = asyncio.get_event_loop()
        except RuntimeError:
            loop = asyncio.new_event_loop()
            asyncio.set_event_loop(loop)

        try:
            response_message = loop.run_until_complete(
                self._process_message(user_id, message)
            )
            return agent_pb2.SendAgentMessageResponse(message=response_message)
        except Exception as e:
            logger.error(
                f"Error processing message from user {user_id}: {str(e)}"
            )
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Error processing message: {str(e)}")
            return agent_pb2.SendAgentMessageResponse()
