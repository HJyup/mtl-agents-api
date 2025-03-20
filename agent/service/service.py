from agent.protos import agent_pb2_grpc

class AgentServicer(agent_pb2_grpc.AgentServiceServicer):
    def CreateAgentStream(self, request, context):
        pass

    def SendAgentMessage(self, request, context):
        pass