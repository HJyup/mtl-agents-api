package service

import (
	"context"
	"encoding/base64"
	"errors"
	pb "github.com/HJyup/mtl-common/api"
	"github.com/HJyup/mtl-common/utils"
	"go.uber.org/zap"
)

var (
	ErrorEmptyUserID          = errors.New("user ID is required")
	ErrorNotFound             = errors.New("configuration not found")
	ErrorInvalidEncryptionKey = errors.New("invalid encryption key length")
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
	encKey []byte
}

func NewService(store Store, logger *zap.Logger, encKeyStr string) (*Service, error) {
	encKey, err := base64.StdEncoding.DecodeString(encKeyStr)
	if err != nil {
		return nil, err
	}

	if len(encKey) != 32 {
		return nil, ErrorInvalidEncryptionKey
	}

	return &Service{
		store:  store,
		logger: logger,
		encKey: encKey,
	}, nil
}

func (svc *Service) CreateConfiguration(ctx context.Context, p *pb.CreateConfigurationRequest) (*pb.CreateConfigurationResponse, error) {
	if p.UserId == "" {
		return nil, ErrorEmptyUserID
	}

	_, err := svc.store.CreateConfiguration(ctx, p.UserId)
	if err != nil {
		svc.logger.Error("failed to create configuration", zap.Error(err), zap.String("userID", p.UserId))
		return nil, err
	}

	return &pb.CreateConfigurationResponse{
		Message: "Configuration created successfully",
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

	openAIKey, err := utils.Decrypt(config.OpenAIKey, svc.encKey)
	if err != nil {
		svc.logger.Error("failed to decrypt OpenAI key", zap.Error(err))
		return nil, errors.New("failed to decrypt API key")
	}

	googleAPIKey, err := utils.Decrypt(config.Calendar.GoogleAPIKey, svc.encKey)
	if err != nil {
		svc.logger.Error("failed to decrypt Google API key", zap.Error(err))
		googleAPIKey = ""
	}

	return &pb.GetConfigurationResponse{
		UserId:    config.UserID,
		OpenAiKey: openAIKey,
		Calendar: &pb.CalendarConfig{
			GoogleApiKey: googleAPIKey,
			Context:      config.Calendar.Context,
		},
		Things: &pb.ThingsConfig{
			Context: config.Things.Context,
		},
	}, nil
}

func (svc *Service) UpdateConfiguration(ctx context.Context, p *pb.UpdateConfigurationRequest) (*pb.UpdateConfigurationResponse, error) {
	if p.UserId == "" {
		return nil, ErrorEmptyUserID
	}

	existingConfig, err := svc.store.GetConfiguration(ctx, p.UserId)
	if err != nil {
		svc.logger.Error("failed to get configuration for update", zap.Error(err), zap.String("userID", p.UserId))
		return nil, err
	}

	if existingConfig == nil {
		return nil, ErrorNotFound
	}

	encryptedOpenAIKey := existingConfig.OpenAIKey
	if p.OpenAiKey != "" {
		encryptedOpenAIKey, err = utils.Encrypt(p.OpenAiKey, svc.encKey)
		if err != nil {
			svc.logger.Error("failed to encrypt OpenAI key", zap.Error(err))
			return nil, errors.New("failed to encrypt API key")
		}
	}

	calendarConfig := existingConfig.Calendar
	thingsConfig := existingConfig.Things

	if p.Calendar != nil {
		if p.Calendar.GoogleApiKey != "" {
			encryptedGoogleAPIKey, err := utils.Encrypt(p.Calendar.GoogleApiKey, svc.encKey)
			if err != nil {
				svc.logger.Error("failed to encrypt Google API key", zap.Error(err))
				encryptedGoogleAPIKey = ""
			}
			calendarConfig.GoogleAPIKey = encryptedGoogleAPIKey
		}
		if p.Calendar.Context != "" {
			calendarConfig.Context = p.Calendar.Context
		}
	}

	if p.Things != nil && p.Things.Context != "" {
		thingsConfig.Context = p.Things.Context
	}

	updatedConfig := &Configuration{
		UserID:    p.UserId,
		OpenAIKey: encryptedOpenAIKey,
		Calendar:  calendarConfig,
		Things:    thingsConfig,
	}

	_, err = svc.store.UpdateConfiguration(ctx, updatedConfig)
	if err != nil {
		svc.logger.Error("failed to update configuration", zap.Error(err), zap.String("userID", p.UserId))
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
