package eino

import (
	"os"

	"gopkg.in/yaml.v3"
)

type AIConfig struct {
	LLMApiKey  string
	LLMBaseURL string
	LLMModel   string
}

func ReadAIConfig() *AIConfig {
	cfg := &AIConfig{
		LLMBaseURL: "https://api.deepseek.com",
		LLMModel:   "deepseek-chat",
	}
	cfg.LLMApiKey = os.Getenv("LLM_API_KEY")
	if cfg.LLMApiKey != "" {
		return cfg
	}
	paths := []string{"configs/config.yaml", "../../configs/config.yaml"}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var raw struct {
			AI *struct {
				LLMApiKey  string `yaml:"llm_api_key"`
				LLMBaseURL string `yaml:"llm_base_url"`
				LLMModel   string `yaml:"llm_model"`
			} `yaml:"ai"`
		}
		if err := yaml.Unmarshal(data, &raw); err != nil || raw.AI == nil {
			continue
		}
		if raw.AI.LLMApiKey != "" {
			cfg.LLMApiKey = raw.AI.LLMApiKey
		}
		if raw.AI.LLMBaseURL != "" {
			cfg.LLMBaseURL = raw.AI.LLMBaseURL
		}
		if raw.AI.LLMModel != "" {
			cfg.LLMModel = raw.AI.LLMModel
		}
		return cfg
	}
	return cfg
}