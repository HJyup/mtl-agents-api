package service

import (
	"context"
	"errors"
	pb "github.com/HJyup/mtl-common/api"
	"go.uber.org/zap"
)

var (
	ErrorEmptyUserID   = errors.New("user ID is required")
	ErrorEmptyConfigID = errors.New("config ID is required")
	ErrorNotFound      = errors.New("configuration not found")
)

type Store interface {
	CreateConfiguration(ctx context.Context, userID string) (*Configuration, error)
	GetConfiguration(ctx context.Context, userID string) (*Configuration, error)
	UpdateConfiguration(ctx context.Context, config *Configuration) (*Configuration, error)
	DeleteConfiguration(ctx context.Context, userID string) error
}

type Service struct {
	store  Store
	logger *zap.Logger
}

func NewService(store Store, logger *zap.Logger) *Service {
	return &Service{store: store, logger: logger}
}

func (svc *Service) CreateConfiguration(ctx context.Context, p *pb.CreateConfigurationRequest) (*pb.CreateConfigurationResponse, error) {
	if p.UserId == "" {
		return nil, ErrorEmptyUserID
	}

	config, err := svc.store.CreateConfiguration(ctx, p.UserId)
	if err != nil {
		svc.logger.Error("failed to create configuration", zap.Error(err), zap.String("userID", p.UserId))
		return nil, err
	}

	return &pb.CreateConfigurationResponse{
		ConfigId: config.ID,
		Message:  "Configuration created successfully",
	}, nil
}

func (svc *Service) GetConfiguration(ctx context.Context, p *pb.GetConfigurationRequest) (*pb.GetConfigurationResponse, error) {
	if p.UserId == "" {
		return nil, ErrorEmptyUserID
	}

	config, err := svc.store.GetConfiguration(ctx, p.UserId)
	if err != nil {
		svc.logger.Error("failed to get configuration", zap.Error(err), zap.String("userID", p.UserId))
		return nil, err
	}

	if config == nil {
		return nil, ErrorNotFound
	}

	calendarConfig, thingsConfig := mapAgentsToProtoConfigs(config.Agents)

	return &pb.GetConfigurationResponse{
		ConfigId:  config.ID,
		UserId:    config.UserID,
		OpenAiKey: config.OpenAIKey,
		Calendar:  calendarConfig,
		Things:    thingsConfig,
	}, nil
}

func (svc *Service) UpdateConfiguration(ctx context.Context, p *pb.UpdateConfigurationRequest) (*pb.UpdateConfigurationResponse, error) {
	if p.ConfigId == "" {
		return nil, ErrorEmptyConfigID
	}

	existingConfig, err := svc.store.GetConfiguration(ctx, p.ConfigId)
	if err != nil {
		svc.logger.Error("failed to get configuration for update", zap.Error(err), zap.String("configID", p.ConfigId))
		return nil, err
	}

	if existingConfig == nil {
		return nil, ErrorNotFound
	}

	agents := mapProtoConfigsToAgents(p.Calendar, p.Things)

	config := &Configuration{
		ID:        p.ConfigId,
		UserID:    existingConfig.UserID,
		OpenAIKey: p.OpenAiKey,
		Agents:    agents,
	}

	_, err = svc.store.UpdateConfiguration(ctx, config)
	if err != nil {
		svc.logger.Error("failed to update configuration", zap.Error(err), zap.String("configID", p.ConfigId))
		return nil, err
	}

	return &pb.UpdateConfigurationResponse{
		Success: true,
		Message: "Configuration updated successfully",
	}, nil
}

func (svc *Service) DeleteConfiguration(ctx context.Context, p *pb.DeleteConfigurationRequest) (*pb.DeleteConfigurationResponse, error) {
	if p.UserId == "" {
		return nil, ErrorEmptyUserID
	}

	err := svc.store.DeleteConfiguration(ctx, p.UserId)
	if err != nil {
		svc.logger.Error("failed to delete configuration", zap.Error(err), zap.String("userID", p.UserId))
		return nil, err
	}

	return &pb.DeleteConfigurationResponse{
		Success: true,
		Message: "Configuration deleted successfully",
	}, nil
}

func mapAgentsToProtoConfigs(agents []Agent) (*pb.CalendarConfig, *pb.ThingsConfig) {
	var calendarConfig *pb.CalendarConfig
	var thingsConfig *pb.ThingsConfig

	for _, agent := range agents {
		switch agent.Type {
		case "calendar":
			calendarConfig = &pb.CalendarConfig{
				GoogleApiKey: agent.GoogleAPIKey,
				Context:      agent.Context,
			}
		case "things":
			thingsConfig = &pb.ThingsConfig{
				Context: agent.Context,
			}
		}
	}

	return calendarConfig, thingsConfig
}

func mapProtoConfigsToAgents(calendarConfig *pb.CalendarConfig, thingsConfig *pb.ThingsConfig) []Agent {
	var agents []Agent

	if calendarConfig != nil {
		agents = append(agents, Agent{
			Type:         "calendar",
			GoogleAPIKey: calendarConfig.GoogleApiKey,
			Context:      calendarConfig.Context,
		})
	}

	if thingsConfig != nil {
		agents = append(agents, Agent{
			Type:    "things",
			Context: thingsConfig.Context,
		})
	}

	return agents
}
