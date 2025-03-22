package models

type CalendarConfig struct {
	GoogleAPIKey string `json:"google_api_key"`
	Context      string `json:"context"`
}

type ThingsConfig struct {
	Context string `json:"context"`
}

type UpdateConfigurationRequest struct {
	OpenAIKey string          `json:"open_ai_key"`
	Calendar  *CalendarConfig `json:"calendar"`
	Things    *ThingsConfig   `json:"things"`
}
