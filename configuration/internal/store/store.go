package store

import (
	"context"
	"errors"
	"fmt"
	"github.com/HJyup/mlt-configuration/internal/service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Store struct {
	client *mongo.Client
}

func NewStore(client *mongo.Client) *Store {
	return &Store{client: client}
}

func (s *Store) getCollection() *mongo.Collection {
	return s.client.Database("mlt-agents-configuration").Collection("configs")
}

func (s *Store) CreateConfiguration(ctx context.Context, userID string) (*service.Configuration, error) {
	collection := s.getCollection()

	var existingConfig service.Configuration
	err := collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&existingConfig)
	if err == nil {
		return nil, errors.New("configuration already exists for this user")
	} else if !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, fmt.Errorf("error checking existing configuration: %w", err)
	}

	config := service.Configuration{
		UserID: userID,
		Agents: []service.Agent{},
	}

	result, err := collection.InsertOne(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create configuration: %w", err)
	}

	config.ID = result.InsertedID.(string)

	return &config, nil
}

func (s *Store) GetConfiguration(ctx context.Context, userID string) (*service.Configuration, error) {
	collection := s.getCollection()
	var config service.Configuration
	err := collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration: %w", err)
	}

	return &config, nil
}

func (s *Store) UpdateConfiguration(ctx context.Context, config *service.Configuration) (*service.Configuration, error) {
	collection := s.getCollection()

	filter := bson.M{"user_id": config.UserID}

	if config.ID == "" {
		filter = bson.M{"_id": config.ID}
	}

	opts := options.FindOneAndReplace().SetReturnDocument(options.After)

	var updatedConfig service.Configuration
	err := collection.FindOneAndReplace(ctx, filter, config, opts).Decode(&updatedConfig)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("configuration not found")
		}
		return nil, fmt.Errorf("failed to update configuration: %w", err)
	}

	return &updatedConfig, nil
}

func (s *Store) DeleteConfiguration(ctx context.Context, userID string) error {
	collection := s.getCollection()

	result, err := collection.DeleteOne(ctx, bson.M{"user_id": userID})
	if err != nil {
		return fmt.Errorf("failed to delete configuration: %w", err)
	}

	if result.DeletedCount == 0 {
		return errors.New("configuration not found")
	}

	return nil
}
