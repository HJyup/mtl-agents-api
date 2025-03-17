from agents import set_default_openai_key
from dotenv import load_dotenv
import asyncio
import os

from service.agent_service import AgentServicer

load_dotenv()

set_default_openai_key(key=os.getenv("OPENAI_API_KEY"))


async def run_agent_via_grpc():
    """Demonstrate how to use the agent via gRPC."""
    print("\nRunning agent via gRPC (simulated)...")

    service = AgentServicer()

    user_id = "user123"
    config_id = "config456"

    responses = [r async for r in service.process_agent_request(user_id, config_id)]
    thread_id = responses[0].thread_id
    print(f"Created thread with ID: {thread_id}")

    message = "Hola, ¿cómo estás?"
    response = await service.process_message(thread_id, user_id, message)
    print(f"Response: {response.message}")


async def main():
    await run_agent_via_grpc()


if __name__ == "__main__":
    asyncio.run(main())