package main

import (
	"context"
	"fmt"
	"github.com/glifery/openai-assistant-with-tools/function_calling/google_search"
	"github.com/glifery/openai-assistant-with-tools/function_calling/web_crawler"
	"github.com/glifery/openai-assistant-with-tools/openai_service"
	"github.com/glifery/openai-assistant-with-tools/types"
	"os/signal"
	"syscall"
	"time"
)

var cfg Config

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	LoadDotenv()
	cfg = LoadConfig(ctx)

	googleSearchFunction := google_search.NewGoogleSearchFunction(ctx, google_search.GoogleSearchConfig{
		ApiKey:         cfg.GoogleSearch.ApiKey,
		SearchEngineID: cfg.GoogleSearch.SearchEngineId,
		MockResponse:   cfg.Function.GoogleSearch.MockResponse,
	})
	webCrawlerFunction := web_crawler.NewWebCrawlerFunction()

	openApiService := openai_service.NewAssistantService(openai_service.AssistantConfig{
		Name: "Google-Search-Test",
		//Instructions: "You are helpful assistant who can answer user's questions. You can search in Google Search using function calling. Also you can parse website to extract content. Also you can find some link on website and pass to the next page and parse it untill you find answer. Combine those skills to give correct answer. If you stuck with google search then modify search query and try again",
		//Instructions: "You are helpful assistant who can answer user's questions. Use Google Search to find some links. Use web crowler to parse HTML of website. Modify search query if can't find an answer. Use links found on a website to deep search.",
		Instructions: "Google Search Integration:\n\nWhen a user asks a question that requires current information or specific knowledge, the assistant should use its Google Search integration.\nThe assistant will send a query related to the user's question to Google Search.\nAfter receiving the list of search results, the assistant should select the most relevant and trustworthy sources to answer the question.\n\nContent Retrieval by URL:\n\nIf the user provides a specific URL or if the assistant identifies a highly relevant source from the search results, the assistant can use its function to retrieve content directly from that URL.\nThis function is particularly useful for accessing detailed information from specific websites or articles.\n\nAnswering the User's Question:\n\nThe assistant should synthesize the information obtained from the search results or the retrieved content to provide a comprehensive and accurate answer to the user's question.\nIt should ensure that the answer is concise, relevant, and easy to understand, avoiding unnecessary technical jargon unless requested by the user.\n\nCiting Sources:\n\nThe assistant should cite its sources clearly, providing the titles of articles or the names of websites from which the information was obtained.\nIf direct quotes are used, they should be clearly indicated.\n\nPrivacy and Discretion:\n\nThe assistant must respect user privacy and confidentiality, not storing or using personal information from the queries beyond the scope of the immediate question.\nIt should avoid retrieving content from websites that are known to be unreliable or that may contain sensitive or controversial information.",
		Token:        cfg.OpenAi.ApiKey,
	}, []types.FunctionCalling{
		googleSearchFunction,
		webCrawlerFunction,
	})

	//err := openApiService.CreateFromScratch()
	err := openApiService.ConnectToExistingThread(ctx, cfg.OpenAi.AssistantId, cfg.OpenAi.ThreadId)
	if err != nil {
		panic(err)
	}

	question := "When was the last strong earthquake in Bali? Get amount of victims."
	//question := "Give short summarize of russian news for the last week."
	//question := "Give a list of upcoming tech companies presentation and their products. Order by date. Specify which product they are planning to present."
	//question := "Give latest leaks about Apple upcoming products"
	//question := "Какие выставки современного искусства работают сейчас в Варшаве?"
	//question := "Расскажи подробнее про фестиваль современного искусства в Варшаве, который стартует на этой неделе"
	//question := "Предостась мне список фильмов, которые сейчас можно посмотреть в Минске. Покажи ближайший сеанс, стоимость билетов, имя режиссера и рейтинг на IMDB. Отсортируй результат по рейтингу от максимального к минимальному"
	//question := "Покажи адреса кинотеатров Минска. Отсортируй их по рейтингу"
	//question := "Какие часы работы выставки \"Пейзаж польской живописи\"?"
	//question := "Is it proper answer to user's question? If not then try again with different query. If yes then say it to user."
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
