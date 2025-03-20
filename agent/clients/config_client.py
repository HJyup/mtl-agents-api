import grpc
from typing import Optional
from agent.protos import config_pb2, config_pb2_grpc
from agent.utils.consul import ConsulClient

class ConfigurationClient:
    
    def __init__(self, consul_client: ConsulClient, service_name: str = "configuration"):
        self.consul_client = consul_client
        self.service_name = service_name
        self.channel = None
        self.stub = None
        
    def connect(self) -> bool:
        try:
            service_info = self.consul_client.discover_service(self.service_name)
            if not service_info:
                print(f"Configuration service '{self.service_name}' not found in Consul")
                return False
                
            host, port = service_info
            target = f"{host}:{port}"

            self.channel = grpc.insecure_channel(target)
            self.stub = config_pb2_grpc.ConfigurationServiceStub(self.channel)
            
            print(f"Connected to configuration service at {target}")
            return True
        except Exception as e:
            print(f"Error connecting to configuration service: {str(e)}")
            return False
    
    def close(self):
        if self.channel:
            self.channel.close()
            
    def get_configuration(self, user_id: str) -> Optional[config_pb2.GetConfigurationResponse]:
        if not self.stub:
            if not self.connect():
                return None
        
        try:
            request = config_pb2.GetConfigurationRequest(user_id=user_id)
            response = self.stub.GetConfigurationByUserID(request)
            return response
        except grpc.RpcError as e:
            status_code = e.code()
            if status_code == grpc.StatusCode.NOT_FOUND:
                print(f"Configuration for user {user_id} not found")
            else:
                print(f"Error getting configuration: {e.details()}")
            return None
        except Exception as e:
            print(f"Unexpected error getting configuration: {str(e)}")
            return None 