import asyncio
import time
import uuid
from concurrent import futures

import grpc
from agents import Agent, Runner

from agent.protos import agent_pb2
from agent.protos import agent_pb2_grpc


class AgentServicer(agent_pb2_grpc.AgentServiceServicer):
    """Implementation of AgentService service."""

    def __init__(self):
        self.active_threads = {}

    async def process_agent_request(self, user_id, config_id):
        agent = Agent(
            name="Default agent",
            instructions="You are a helpful assistant.",
        )

        thread_id = str(uuid.uuid4())

        yield agent_pb2.AgentStreamResponse(
            thread_id=thread_id,
            message="Agent initialized. Ready to assist."
        )

        self.active_threads[thread_id] = {
            "agent": agent,
            "user_id": user_id,
            "config_id": config_id
        }

    def CreateAgentStream(self, request, context):
        """Implementation of CreateAgentStream RPC method."""
        user_id = request.user_id
        config_id = request.config_id

        for response in asyncio.run(self.process_agent_request(user_id, config_id)):
            yield response

    async def process_message(self, thread_id, message):
        """Process user message in a specific thread."""
        if thread_id not in self.active_threads:
            return agent_pb2.SendAgentMessageResponse(
                thread_id=thread_id,
                message="Error: Thread not found"
            )

        thread_data = self.active_threads[thread_id]
        agent = thread_data["agent"]

        try:
            result = await Runner.run(agent, input=message)
            return agent_pb2.SendAgentMessageResponse(
                thread_id=thread_id,
                message=result.final_output
            )
        except Exception as e:
            return agent_pb2.SendAgentMessageResponse(
                thread_id=thread_id,
                message=f"Error processing message: {str(e)}"
            )

    def SendAgentMessage(self, request, context):
        """Implementation of SendAgentMessage RPC method."""
        thread_id = request.thread_id
        user_id = request.user_id
        message = request.message

        return asyncio.run(self.process_message(thread_id, message))


def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    agent_pb2_grpc.add_AgentServiceServicer_to_server(AgentServicer(), server)
    server.add_insecure_port('[::]:50051')
    server.start()
    print("Agent service started on port 50051")
    try:
        while True:
            time.sleep(86400)
    except KeyboardInterrupt:
        server.stop(0)
        print("Server stopped")


if __name__ == '__main__':
    serve() 