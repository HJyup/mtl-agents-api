import grpc
import logging
from typing import Optional
from agent.protos import config_pb2, config_pb2_grpc
from agent.utils.consul import ConsulClient

logging.basicConfig(level=logging.INFO)


class ConfigurationClient:
    def __init__(self, consul_client: ConsulClient, service_name: str = "configuration"):
        self.consul_client = consul_client
        self.service_name = service_name
        self.channel = None
        self.stub = None

    def connect(self) -> bool:
        try:
            service_info = self.consul_client.discover(self.service_name)
            if not service_info:
                logging.error("Service '%s' not found in Consul.", self.service_name)
                return False

            port = service_info[0]
            self.channel = grpc.insecure_channel(port)
            self.stub = config_pb2_grpc.ConfigurationServiceStub(self.channel)
            logging.info("Successfully connected to service '%s' on port %s.",
                         self.service_name, port)
            return True

        except grpc.RpcError as e:
            logging.error("gRPC error during connect: %s", e)
            return False

        except Exception as e:
            logging.error("Unexpected error during connect: %s", e)
            return False

    def get_configuration(self, user_id: str) -> Optional[config_pb2.GetConfigurationResponse]:
        if not self.stub and not self.connect():
            return None

        try:
            request = config_pb2.GetConfigurationRequest(user_id=user_id)
            response = self.stub.GetConfigurationByUserID(request)
            logging.info("Configuration retrieved for user_id: %s", user_id)
            return response

        except grpc.RpcError as rpc_error:
            logging.error("gRPC error in get_configuration: %s", rpc_error)

        except Exception as error:
            logging.error("Unexpected error in get_configuration: %s", error)

        return None
