#!/usr/bin/env python3
import os
import time
import consul
import logging
from typing import List

logger = logging.getLogger(__name__)


class ConsulClient:
    def __init__(self, addr: str):
        host, port = addr.split(":")
        self.client = consul.Consul(host=host, port=int(port))

    def register(self, instance_id: str, server_name: str, host_port: str) -> bool:
        host, port = host_port.split(":")
        port = int(port)
        self.client.agent.service.register(
            name=server_name,
            service_id=instance_id,
            address=host,
            port=port,
            check=consul.Check.ttl("10s"),
        )
        return True

    def deregister(self, instance_id: str) -> bool:
        self.client.agent.service.deregister(instance_id)
        return True

    def discover(self, server_name: str) -> List[str]:
        _, services = self.client.health.service(server_name, passing=True)
        return [
            f"{service['Service']['Address']}:{service['Service']['Port']}"
            for service in services
        ]

    def health_check(self, instance_id: str) -> bool:
        self.client.agent.check.ttl_pass(f"service:{instance_id}", "online")
        return True


def health_check_loop(consul_client: ConsulClient, instance_id: str):
    while True:
        try:
            consul_client.health_check(instance_id)
            time.sleep(3)
        except Exception as e:
            logger.error(f"Health check failed: {e}")
            break


def generate_instance_id(server_name: str) -> str:
    return f"{server_name}-{os.urandom(8).hex()}"
