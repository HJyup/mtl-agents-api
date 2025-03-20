import logging
import uuid
import grpc
from concurrent.futures import ThreadPoolExecutor
from threading import Lock

from agent.clients import ConfigurationClient
from agent.protos import agent_pb2, agent_pb2_grpc

logger = logging.getLogger(__name__)

class AgentServicer(agent_pb2_grpc.AgentServiceServicer):
    def __init__(self, configuration_service):
        self.configuration_service: ConfigurationClient = configuration_service
        self.active_conversations = {}
        self.lock = Lock()
        self.executor = ThreadPoolExecutor(max_workers=10)

    def CreateAgentStream(self, request, context):
        user_id = request.user_id
        thread_id = str(uuid.uuid4())

        configuration = self.configuration_service.get_configuration(user_id)

        with self.lock:
            self.active_conversations[user_id] = thread_id

        logger.info(f"Created new conversation {thread_id} for user {user_id}")

        initial_response = agent_pb2.AgentStreamResponse(
            thread_id=thread_id,
            message="Agent conversation initialized"
        )
        yield initial_response

        try:
            context.add_callback(lambda: logger.info(f"Stream closed for conversation {thread_id}"))
            while not context.is_active() or context.cancelled():
                pass
        except Exception as e:
            logger.error(f"Error in stream for conversation {thread_id}: {str(e)}")
        finally:
            with self.lock:
                if user_id in self.active_conversations and self.active_conversations[user_id] == thread_id:
                    del self.active_conversations[user_id]
                    logger.info(f"Conversation {thread_id} for user {user_id} cleaned up")

    def SendAgentMessage(self, request, context):
        thread_id = request.thread_id
        user_id = request.user_id
        message = request.message

        logger.info(f"Received message for conversation {thread_id} from user {user_id}")

        with self.lock:
            if user_id not in self.active_conversations:
                context.set_code(grpc.StatusCode.NOT_FOUND)
                context.set_details(f"No active conversation for user {user_id}")
                return agent_pb2.SendAgentMessageResponse()

            if self.active_conversations[user_id] != thread_id:
                context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
                context.set_details(f"Thread ID {thread_id} does not match active conversation for user {user_id}")
                return agent_pb2.SendAgentMessageResponse()

        response_message = f"Processed: {message}"

        return agent_pb2.SendAgentMessageResponse(
            message=response_message
        )