package gateway

import (
	"context"
	common "github.com/HJyup/mtl-common"
	pb "github.com/HJyup/mtl-common/api"
	"go.uber.org/zap"
)

var (
	UserServiceName          = "user"
	FailedToConnectUserError = "Failed to connect to chat service"
)

type UserGateway struct {
	registry common.Registry
	logger   *zap.Logger
}

func NewUserGateway(registry common.Registry, logger *zap.Logger) *UserGateway {
	return &UserGateway{registry: registry, logger: logger}
}

func (g *UserGateway) CreatUser(ctx context.Context, payload *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	conn, err := common.ServiceConnection(context.Background(), UserServiceName, g.registry)
	if err != nil {
		g.logger.Error(FailedToConnectUserError)
	}
	chatClient := pb.NewUserServiceClient(conn)
	return chatClient.CreateUser(ctx, payload)
}

func (g *UserGateway) AuthUser(ctx context.Context, payload *pb.AuthUserRequest) (*pb.AuthUserResponse, error) {
	conn, err := common.ServiceConnection(context.Background(), UserServiceName, g.registry)
	if err != nil {
		g.logger.Error(FailedToConnectUserError)
	}
	chatClient := pb.NewUserServiceClient(conn)
	return chatClient.AuthUser(ctx, payload)
}

func (g *UserGateway) GetUser(ctx context.Context, payload *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	conn, err := common.ServiceConnection(context.Background(), UserServiceName, g.registry)
	if err != nil {
		g.logger.Error(FailedToConnectUserError)
	}
	chatClient := pb.NewUserServiceClient(conn)
	return chatClient.GetUser(ctx, payload)
}

func (g *UserGateway) DeleteUser(ctx context.Context, payload *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	conn, err := common.ServiceConnection(context.Background(), UserServiceName, g.registry)
	if err != nil {
		g.logger.Error(FailedToConnectUserError)
	}
	chatClient := pb.NewUserServiceClient(conn)
	return chatClient.DeleteUser(ctx, payload)
}
