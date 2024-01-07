package main

import (
	"context"
	"fmt"
	"google_search_api/function_calling/google_search"
	"google_search_api/openai_service"
	"google_search_api/types"
	"os/signal"
	"syscall"
	"time"
)

var cfg Config

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	LoadDotenv()
	cfg = LoadConfig(ctx)

	googleSearch := google_search.NewGoogleSearchFunction(ctx, google_search.GoogleSearchConfig{
		ApiKey:         cfg.GoogleSearch.ApiKey,
		SearchEngineID: cfg.GoogleSearch.SearchEngineId,
	})
	googleSearch = googleSearch

	openApiService := openai_service.NewAssistantService(openai_service.AssistantConfig{
		Name:         "Google-Search-Test",
		Instructions: "You are helpful assistant to give answers on user's questions. You are able to search in internet using Google Search integration.",
		Token:        cfg.OpenAi.ApiKey,
	}, []types.FunctionCalling{googleSearch})

	//err := openApiService.CreateFromScratch()
	err := openApiService.ConnectToExistingThread(ctx, cfg.OpenAi.AssistantId, cfg.OpenAi.ThreadId)
	if err != nil {
		panic(err)
	}

	question := "When was the last strong earthquake in Bali?"
	//question := "Give short summarize of russian news for the last week."
	//question := "Give a list of upcoming tech companies presentation and their products. Order by date. Specify which product they are planning to present."
	//question := "Give latest leaks about Apple upcoming products"
	run, err := openApiService.SendMessageAndRun(ctx, question)
	if err != nil {
		panic(err)
	}

	messages, err := openApiService.GetRunResponse(ctx, &run)
	if err != nil {
		panic(err)
	}

	for _, message := range messages.Messages {
		t := time.Unix(int64(message.CreatedAt), 0)

		fmt.Printf(
			"%s (%s): %s\n",
			message.Role,
			t.Format(time.RFC3339),
			message.Content[0].Text.Value,
		)
	}

	return
}
