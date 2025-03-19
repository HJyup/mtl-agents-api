package store

import "go.mongodb.org/mongo-driver/bson/primitive"

type Agent struct {
	Type         string `bson:"type"`
	GoogleAPIKey string `bson:"google_api_key,omitempty"`
	Context      string `bson:"context,omitempty"`
}

type Configuration struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    int                `bson:"user_id"`
	OpenAIKey string             `bson:"openai_key,omitempty"`
	Agents    []Agent            `bson:"agents,omitempty"`
}
