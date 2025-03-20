import asyncio
import time
import uuid
from concurrent import futures
import logging
import grpc

from agent.protos import agent_pb2, agent_pb2_grpc
from agent.clients import ConfigurationClient
from agents import Agent, Runner, set_default_openai_key

logger = logging.getLogger(__name__)


class AgentServicer(agent_pb2_grpc.AgentServiceServicer):

    def __init__(self, config_client: ConfigurationClient):
        self.active_threads = {}
        self.config_client = config_client
        
    async def process_agent_request(self, user_id: str, config_id: str):
        logger.info(f"Creating agent thread for user {user_id} with config {config_id}")

        config = self.config_client.get_configuration(user_id)

        instructions = "You are a helpful assistant."
        name = "Default Agent"

        if config:
            logger.info(f"Found configuration for user {user_id}")
            if config.open_ai_key:
                try:
                    set_default_openai_key(key=config.open_ai_key)
                    logger.info("Set OpenAI API key from config")
                except Exception as e:
                    logger.error(f"Error decrypting OpenAI key: {e}")
        else:
            logger.warning(f"No configuration found for user {user_id}")

        agent = Agent(
            name=name,
            instructions=instructions,
        )

        thread_id = str(uuid.uuid4())

        self.active_threads[thread_id] = {
            "agent": agent,
            "user_id": user_id,
            "config_id": config_id,
            "created_at": time.time()
        }

        yield agent_pb2.AgentStreamResponse(
            thread_id=thread_id,
            message="Agent initialized. Ready to assist."
        )

    def CreateAgentStream(self, request, context):
        """Implementation of CreateAgentStream RPC method."""
        user_id = request.user_id
        config_id = request.config_id
        
        logger.info(f"CreateAgentStream called for user {user_id}")

        context.set_compression(grpc.Compression.Gzip)

        for response in asyncio.run(self.process_agent_request(user_id, config_id)):
            yield response

    async def process_message(self, thread_id: str, user_id: str, message: str):
        logger.info(f"Processing message in thread {thread_id} for user {user_id}")

        if thread_id not in self.active_threads:
            logger.warning(f"Thread {thread_id} not found")
            return agent_pb2.SendAgentMessageResponse(
                thread_id=thread_id,
                message="Error: Thread not found"
            )

        thread_data = self.active_threads[thread_id]

        if thread_data["user_id"] != user_id:
            logger.warning(f"User {user_id} attempted to access thread {thread_id} belonging to {thread_data['user_id']}")
            return agent_pb2.SendAgentMessageResponse(
                thread_id=thread_id,
                message="Error: Unauthorized access to thread"
            )
        
        agent = thread_data["agent"]

        try:
            logger.info(f"Running agent for message: {message[:50]}...")
            result = await Runner.run(agent, input=message)

            thread_data["last_activity"] = time.time()
            
            return agent_pb2.SendAgentMessageResponse(
                thread_id=thread_id,
                message=result.final_output
            )
        except Exception as e:
            logger.error(f"Error processing message: {e}", exc_info=True)
            return agent_pb2.SendAgentMessageResponse(
                thread_id=thread_id,
                message=f"Error processing message: {str(e)}"
            )

    def SendAgentMessage(self, request, context):
        thread_id = request.thread_id
        user_id = request.user_id
        message = request.message
        
        logger.info(f"SendAgentMessage called for thread {thread_id}")
        
        return asyncio.run(self.process_message(thread_id, user_id, message))


class AgentServer:
    
    def __init__(self, host: str, port: int, config_client: ConfigurationClient):
        self.host = host
        self.port = port
        self.server = None
        self.config_client = config_client
        
    def start(self):
        """Start the gRPC server."""
        self.server = grpc.server(
            futures.ThreadPoolExecutor(max_workers=10),
            compression=grpc.Compression.Gzip
        )

        servicer = AgentServicer(self.config_client)
        agent_pb2_grpc.add_AgentServiceServicer_to_server(servicer, self.server)

        address = f"{self.host}:{self.port}"
        self.server.add_insecure_port(address)

        self.server.start()
        logger.info(f"Agent service started on {address}")
        
        return self.server
        
    def stop(self, grace: int = 0):
        if self.server:
            self.server.stop(grace)
            logger.info("Agent service stopped")


def serve(host: str = "0.0.0.0", port: int = 50051, config_client: ConfigurationClient = None):
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )
    
    server = AgentServer(host, port, config_client)
    _ = server.start()
    
    try:
        while True:
            time.sleep(86400)
    except KeyboardInterrupt:
        server.stop(0) 