import logging
import os
import threading
import time
from dotenv import load_dotenv

from agent.service import serve
from agent.utils import ConsulClient, health_check_loop, generate_instance_id

load_dotenv()

logging.basicConfig(level=logging.INFO,
                    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

def main():
    consul_addr = os.getenv("AGENT_CONSUL")
    consul_client = ConsulClient(consul_addr)

    service_name = os.getenv("AGENT_SERVICENAME")
    instance_id = generate_instance_id(service_name)
    host_port = os.getenv("AGENT_ADDRESS")

    try:
        consul_client.register(instance_id, service_name, host_port)
        server = serve(host_port, consul_client)
        logging.info(f"Service registered: {service_name} (ID: {instance_id})")

        health_thread = threading.Thread(
            target=health_check_loop,
            args=(consul_client, instance_id),
            daemon=True
        )
        health_thread.start()
        logging.info("Service is running...")

        try:
            while True:
                time.sleep(1)
        except KeyboardInterrupt:
            logging.info("Shutting down...")
            server.stop(0)

    finally:
        consul_client.deregister(instance_id)
        logging.info(f"Service deregistered: {instance_id}")


if __name__ == "__main__":
    main()