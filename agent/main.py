import logging
from dotenv import load_dotenv
from agent.config import settings
from agent.utils.consul import ConsulClient
from agent.clients.config_client import ConfigurationClient
from agent.service.agent_service import serve as start_agent_service
from agents import set_default_openai_key

load_dotenv()

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

def main():
    """Main entry point for the agent service."""
    logger.info("Starting agent service...")

    if settings.OPENAI_API_KEY:
        set_default_openai_key(key=settings.OPENAI_API_KEY)
        logger.info("Set OpenAI API key from environment")

    consul_client = ConsulClient(
        host=settings.CONSUL_HOST,
        port=settings.CONSUL_PORT,
        service_name=settings.SERVICE_NAME,
        service_port=settings.SERVICE_PORT,
        service_host=settings.SERVICE_HOST
    )

    try:
        instance_id = consul_client.register()
        logger.info(f"Registered with Consul as {instance_id}")
    except Exception as e:
        logger.error(f"Failed to register with Consul: {e}")
        logger.warning("Running without service discovery")

    config_client = ConfigurationClient(
        consul_client=consul_client,
        service_name=settings.CONFIG_SERVICE_NAME
    )
    
    try:
        start_agent_service(
            host=settings.SERVICE_HOST,
            port=settings.SERVICE_PORT,
            config_client=config_client,
        )
    except KeyboardInterrupt:
        logger.info("Shutting down...")
    except Exception as e:
        logger.error(f"Error starting service: {e}", exc_info=True)
    finally:
        # Deregister from Consul
        try:
            consul_client.deregister()
            logger.info("Deregistered from Consul")
        except Exception as e:
            logger.error(f"Error deregistering from Consul: {e}")


if __name__ == "__main__":
    main()