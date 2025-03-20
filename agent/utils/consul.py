import consul
import socket
import uuid
import time
import threading
from typing import Optional

class ConsulClient:

    def __init__(self, host: str, port: int, service_name: str, service_port: int, service_host: str = "0.0.0.0"):
        self.client = consul.Consul(host=host, port=port)
        self.service_name = service_name
        self.service_port = service_port
        self.service_host = service_host
        self.instance_id = f"{service_name}-{uuid.uuid4()}"
        self._health_check_thread = None
        self._stop_event = threading.Event()

    def register(self):
        """Register the service with Consul."""
        self.client.agent.service.register(
            name=self.service_name,
            service_id=self.instance_id,
            address=self.service_host,
            port=self.service_port,
            check=consul.Check.tcp(self.service_host, self.service_port, "10s")
        )
        print(f"Registered service {self.service_name} with ID {self.instance_id}")

        self._health_check_thread = threading.Thread(target=self._health_check_loop)
        self._health_check_thread.daemon = True
        self._health_check_thread.start()
        
        return self.instance_id

    def deregister(self):
        self._stop_event.set()
        if self._health_check_thread and self._health_check_thread.is_alive():
            self._health_check_thread.join(timeout=1)
            
        self.client.agent.service.deregister(self.instance_id)
        print(f"Deregistered service {self.service_name} with ID {self.instance_id}")

    def _health_check_loop(self):
        while not self._stop_event.is_set():
            try:
                self.client.agent.check.ttl_pass(f"service:{self.instance_id}")
                time.sleep(5)
            except Exception as e:
                print(f"Error in health check: {str(e)}")
                time.sleep(1)

    def discover_service(self, service_name: str) -> Optional[tuple[str, int]]:
        index, services = self.client.catalog.service(service_name)
        if not services:
            return None

        service = services[0]
        return service["ServiceAddress"], service["ServicePort"]
        
    @staticmethod
    def generate_instance_id(service_name: str) -> str:
        """Generate a unique instance ID for a service."""
        hostname = socket.gethostname()
        return f"{service_name}-{hostname}-{uuid.uuid4()}" 