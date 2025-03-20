from agent.service.service import AgentServicer
from agent.protos import agent_pb2_grpc
import grpc
from concurrent import futures

def serve(port):
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    agent_pb2_grpc.add_AgentServiceServicer_to_server(AgentServicer(), server)
    server.add_insecure_port(f'[::]:{port}')
    server.start()
    return server