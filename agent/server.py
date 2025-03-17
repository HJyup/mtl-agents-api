import os
from dotenv import load_dotenv
from agents import set_default_openai_key
from service.agent_service import serve

# Load environment variables
load_dotenv()

# Set OpenAI API key from environment
if os.getenv("OPENAI_API_KEY"):
    set_default_openai_key(key=os.getenv("OPENAI_API_KEY"))
else:
    print("Warning: OPENAI_API_KEY not found in environment variables")

if __name__ == "__main__":
    print("Starting Agent gRPC server...")
    serve() 