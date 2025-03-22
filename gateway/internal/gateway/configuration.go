package gateway

import (
	"context"
	common "github.com/HJyup/mtl-common"
	pb "github.com/HJyup/mtl-common/api"
	"go.uber.org/zap"
)

type Configuration interface {
	CreateConfiguration(context.Context, *pb.CreateConfigurationRequest) (*pb.CreateConfigurationResponse, error)
	UpdateConfiguration(context.Context, *pb.UpdateConfigurationRequest) (*pb.UpdateConfigurationResponse, error)
	GetConfiguration(context.Context, *pb.GetConfigurationRequest) (*pb.GetConfigurationResponse, error)
	DeleteConfiguration(context.Context, *pb.DeleteConfigurationRequest) (*pb.DeleteConfigurationResponse, error)
}

var (
	ConfigurationServiceName          = "user"
	FailedToConnectConfigurationError = "Failed to connect to configuration service"
)

type ConfigurationGateway struct {
	registry common.Registry
	logger   *zap.Logger
}

func NewConfigurationGateway(registry common.Registry, logger *zap.Logger) *ConfigurationGateway {
	return &ConfigurationGateway{registry: registry, logger: logger}
}

func (g *ConfigurationGateway) CreateConfiguration(ctx context.Context, payload *pb.CreateConfigurationRequest) (*pb.CreateConfigurationResponse, error) {
	conn, err := common.ServiceConnection(context.Background(), ConfigurationServiceName, g.registry)
	if err != nil {
		g.logger.Error(FailedToConnectConfigurationError)
	}
	configClient := pb.NewConfigurationServiceClient(conn)
	return configClient.CreateConfiguration(ctx, payload)
}

func (g *ConfigurationGateway) UpdateConfiguration(ctx context.Context, payload *pb.UpdateConfigurationRequest) (*pb.UpdateConfigurationResponse, error) {
	conn, err := common.ServiceConnection(context.Background(), ConfigurationServiceName, g.registry)
	if err != nil {
		g.logger.Error(FailedToConnectConfigurationError)
	}
	configClient := pb.NewConfigurationServiceClient(conn)
	return configClient.UpdateConfiguration(ctx, payload)
}

func (g *ConfigurationGateway) GetConfiguration(ctx context.Context, payload *pb.GetConfigurationRequest) (*pb.GetConfigurationResponse, error) {
	conn, err := common.ServiceConnection(context.Background(), ConfigurationServiceName, g.registry)
	if err != nil {
		g.logger.Error(FailedToConnectConfigurationError)
	}
	configClient := pb.NewConfigurationServiceClient(conn)
	return configClient.GetConfigurationByUserID(ctx, payload)
}

func (g *ConfigurationGateway) DeleteConfiguration(ctx context.Context, payload *pb.DeleteConfigurationRequest) (*pb.DeleteConfigurationResponse, error) {
	conn, err := common.ServiceConnection(context.Background(), ConfigurationServiceName, g.registry)
	if err != nil {
		g.logger.Error(FailedToConnectConfigurationError)
	}
	configClient := pb.NewConfigurationServiceClient(conn)
	return configClient.DeleteConfigurationByUserID(ctx, payload)
}
