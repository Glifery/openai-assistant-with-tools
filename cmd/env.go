package main

import (
	"context"
	"github.com/caarlos0/env/v7"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	OpenAi struct {
		ApiKey      string `env:"OPEN_AI_API_KEY,required"`
		AssistantId string `env:"OPEN_AI_ASSISTANT_ID"`
		ThreadId    string `env:"OPEN_AI_THREAD_ID"`
	}
	GoogleSearch struct {
		ApiKey         string `env:"GOOGLE_SEARCH_API_KEY,required"`
		SearchEngineId string `env:"GOOGLE_SEARCH_SEARCH_ENGINE_ID,required"`
	}
	Function struct {
		GoogleSearch struct {
			MockResponse string `env:"FUNCTION_GOOGLE_SEARCH_MOCK_RESPONSE"`
		}
	}
}

func LoadConfig(_ context.Context) Config {
	cfg := Config{}
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}

func LoadDotenv() {
	err := godotenv.Load()
	if err != nil {
		log.Infof("dotenv file was not loaded: %s", err)
	}
}
