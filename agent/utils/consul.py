import os
import time

import consul
import logging
from typing import List
import contextlib

class ConsulClient:
    def __init__(self, addr: str):
        host, port = addr.split(":")
        self.client = consul.Consul(host=host, port=int(port))
        self.logger = logging.getLogger(__name__)

    def register(self, instance_id: str, server_name: str, host_port: str) -> bool:
        try:
            host, port = host_port.split(":")
            try:
                port = int(port)
            except ValueError:
                raise ValueError("Invalid port")

            self.client.agent.service.register(
                name=server_name,
                service_id=instance_id,
                address=host,
                port=port,
                check=consul.Check.ttl("10s")
            )
            return True
        except Exception as e:
            self.logger.error(f"Failed to register service: {e}")
            raise

    def deregister(self, instance_id: str) -> bool:
        self.logger.info(f"DeRegistering service with ID: {instance_id}")
        try:
            self.client.agent.service.deregister(instance_id)
            return True
        except Exception as e:
            self.logger.error(f"Failed to deregister service: {e}")
            raise

    def discover(self, server_name: str) -> List[str]:
        try:
            _, services = self.client.health.service(server_name, passing=True)
            instances = []

            for service in services:
                address = service['Service']['Address']
                port = service['Service']['Port']
                instances.append(f"{address}:{port}")

            return instances
        except Exception as e:
            self.logger.error(f"Failed to discover services: {e}")
            raise

    def health_check(self, instance_id: str) -> bool:
        try:
            self.client.agent.check.ttl_pass(f"service:{instance_id}", "online")
            return True
        except Exception as e:
            self.logger.error(f"Failed to update health check: {e}")
            raise

@contextlib.contextmanager
def registered_service(client, instance_id, server_name, host_port):
    try:
        client.register(instance_id, server_name, host_port)
        yield
    finally:
        client.deregister(instance_id)


def health_check_loop(consul_client, instance_id):
    while True:
        try:
            consul_client.health_check(instance_id)
            time.sleep(3)
        except Exception as e:
            logging.error(f"Health check failed: {e}")
            break

def generate_instance_id(server_name):
    return f"{server_name}-{os.urandom(8).hex()}"
