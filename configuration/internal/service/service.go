package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	pb "github.com/HJyup/mtl-common/api"
	"go.uber.org/zap"
	"io"
)

var (
	ErrorEmptyUserID          = errors.New("user ID is required")
	ErrorEmptyConfigID        = errors.New("config ID is required")
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

	openAIKey, err := svc.decrypt(config.OpenAIKey)
	if err != nil {
		svc.logger.Error("failed to decrypt OpenAI key", zap.Error(err))
		return nil, errors.New("failed to decrypt API key")
	}

	calendarConfig, thingsConfig := svc.mapAgentsToProtoConfigs(config.Agents)

	return &pb.GetConfigurationResponse{
		ConfigId:  config.ID,
		UserId:    config.UserID,
		OpenAiKey: openAIKey,
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

	encryptedOpenAIKey, err := svc.encrypt(p.OpenAiKey)
	if err != nil {
		svc.logger.Error("failed to encrypt OpenAI key", zap.Error(err))
		return nil, errors.New("failed to encrypt API key")
	}

	agents := svc.mapProtoConfigsToAgents(p.Calendar, p.Things)

	config := &Configuration{
		ID:        p.ConfigId,
		UserID:    existingConfig.UserID,
		OpenAIKey: encryptedOpenAIKey,
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

func (svc *Service) mapAgentsToProtoConfigs(agents []Agent) (*pb.CalendarConfig, *pb.ThingsConfig) {
	var calendarConfig *pb.CalendarConfig
	var thingsConfig *pb.ThingsConfig

	for _, agent := range agents {
		switch agent.Type {
		case "calendar":
			googleAPIKey, err := svc.decrypt(agent.GoogleAPIKey)
			if err != nil {
				svc.logger.Error("failed to decrypt Google API key", zap.Error(err))
				googleAPIKey = ""
			}

			calendarConfig = &pb.CalendarConfig{
				GoogleApiKey: googleAPIKey,
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

func (svc *Service) mapProtoConfigsToAgents(calendarConfig *pb.CalendarConfig, thingsConfig *pb.ThingsConfig) []Agent {
	var agents []Agent

	if calendarConfig != nil {
		encryptedGoogleAPIKey, err := svc.encrypt(calendarConfig.GoogleApiKey)
		if err != nil {
			svc.logger.Error("failed to encrypt Google API key", zap.Error(err))
			encryptedGoogleAPIKey = ""
		}

		agents = append(agents, Agent{
			Type:         "calendar",
			GoogleAPIKey: encryptedGoogleAPIKey,
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

func (svc *Service) decrypt(encrypted string) (string, error) {
	if encrypted == "" {
		return "", nil
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(svc.encKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func (svc *Service) encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	block, err := aes.NewCipher(svc.encKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}
