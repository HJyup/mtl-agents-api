package service

type Configuration struct {
	UserID    string         `bson:"user_id"`
	OpenAIKey string         `bson:"open_ai_key"`
	Calendar  CalendarConfig `bson:"calendar"`
	Things    ThingsConfig   `bson:"things"`
}

type CalendarConfig struct {
	GoogleAPIKey string `bson:"google_api_key"`
	Context      string `bson:"context"`
}

type ThingsConfig struct {
	Context string `bson:"context"`
}
