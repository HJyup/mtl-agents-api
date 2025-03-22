package gateway

import (
	"context"
	common "github.com/HJyup/mtl-common"
	pb "github.com/HJyup/mtl-common/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Agent interface {
	AgentWebsocketStream(context.Context, *pb.AgentMessage) (pb.AgentService_AgentWebsocketStreamClient, error)
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

func (g *ConfigurationGateway) AgentWebsocketStream(ctx context.Context, opts ...grpc.CallOption) (pb.AgentService_AgentWebsocketStreamClient, error) {
	conn, err := common.ServiceConnection(context.Background(), AgentServiceName, g.registry)
	if err != nil {
		g.logger.Error(FailedToConnectAgentError)
	}
	agentClient := pb.NewAgentServiceClient(conn)
	return agentClient.AgentWebsocketStream(ctx, opts...)
}
