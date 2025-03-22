from agent.clients import ConfigurationClient
from agent.service.service import AgentServicer
from agent.protos import agent_pb2_grpc
import grpc
from concurrent import futures

def serve(address, consul_client):
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))

    configuration_service = ConfigurationClient(consul_client)
    agent_pb2_grpc.add_AgentServiceServicer_to_server(AgentServicer(configuration_service), server)

    server.add_insecure_port(address)
    server.start()

    return server