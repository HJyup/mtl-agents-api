package gateway

import (
	"context"
	common "github.com/HJyup/mtl-common"
	pb "github.com/HJyup/mtl-common/api"
	"go.uber.org/zap"
)

type Agent interface {
	CreateAgentStream(context.Context, *pb.CreateAgentStreamRequest) (*pb.AgentStreamResponse, error)
	SendAgentMessage(context.Context, *pb.SendAgentMessageRequest) (*pb.SendAgentMessageResponse, error)
}

var (
	AgentServiceName          = "agent"
	FailedToConnectAgentError = "Failed to connect to agent service"
)

type AgentGateway struct {
	registry common.Registry
	logger   *zap.Logger
}

func NewAgentGateway(registry common.Registry, logger *zap.Logger) *ConfigurationGateway {
	return &ConfigurationGateway{registry: registry, logger: logger}
}

func (g *ConfigurationGateway) CreateAgentStream(ctx context.Context, payload *pb.CreateAgentStreamRequest) (pb.AgentService_CreateAgentStreamClient, error) {
	conn, err := common.ServiceConnection(context.Background(), AgentServiceName, g.registry)
	if err != nil {
		g.logger.Error(FailedToConnectAgentError)
	}
	agentClient := pb.NewAgentServiceClient(conn)
	return agentClient.CreateAgentStream(ctx, payload)
}

func (g *ConfigurationGateway) SendAgentMessage(ctx context.Context, payload *pb.SendAgentMessageRequest) (*pb.SendAgentMessageResponse, error) {
	conn, err := common.ServiceConnection(context.Background(), AgentServiceName, g.registry)
	if err != nil {
		g.logger.Error(FailedToConnectAgentError)
	}
	agentClient := pb.NewAgentServiceClient(conn)
	return agentClient.SendAgentMessage(ctx, payload)
}
