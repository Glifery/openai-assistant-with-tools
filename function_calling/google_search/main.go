package google_search

import (
	"context"
	"encoding/json"
	"github.com/glifery/openai-assistant-with-tools/types"
	"google.golang.org/api/customsearch/v1"
	"google.golang.org/api/option"
	"strings"
)

type GoogleSearchConfig struct {
	ApiKey         string
	SearchEngineID string
	MockResponse   string
}

type GoogleSearchRequest struct {
	Query string `json:"query"`
}

type GoogleSearchResult struct {
	Id      string `json:"id"`
	Link    string `json:"link"`
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
}

type googleSearchFunction struct {
	config  GoogleSearchConfig
	service *customsearch.Service
}

func NewGoogleSearchFunction(ctx context.Context, config GoogleSearchConfig) types.FunctionCalling {
	service, err := customsearch.NewService(ctx, option.WithAPIKey(config.ApiKey))
	if err != nil {
		panic(err)
	}

	return &googleSearchFunction{
		config:  config,
		service: service,
	}
}

func (srv *googleSearchFunction) GetFunctionName() string {
	return "google-search-function"
}

func (srv *googleSearchFunction) GetFunctionDescription() string {
	return "This function searches for a query through Google's search engine using a custom search configuration. It returns list of search results. Each search result contains id, link, title and snippet."
}

func (srv *googleSearchFunction) GetFunctionProperties() map[string]map[string]interface{} {
	return map[string]map[string]interface{}{
		"query": {
			"type":        "string",
			"description": "The search term that the user wants to look up. It is a required field.",
		},
	}
}

func (srv *googleSearchFunction) GetFunctionRequiredProperties() []string {
	return []string{"query"}
}

func (srv *googleSearchFunction) Execute(input string) (output any, err error) {
	if srv.config.MockResponse != "" {
		return srv.config.MockResponse, nil
	}

	var request GoogleSearchRequest
	err = json.Unmarshal([]byte(input), &request)
	if err != nil {
		return nil, err
	}

	results, err := srv.Search(request.Query)
	if err != nil {
		return nil, err
	}

	output, err = srv.Stringify(results)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (srv *googleSearchFunction) Search(query string) ([]GoogleSearchResult, error) {
	f := srv.service.Cse.List().Cx(srv.config.SearchEngineID).Q(query)
	res, err := f.Do()
	if err != nil {
		return nil, err
	}

	results := []GoogleSearchResult{}

	for _, item := range res.Items {
		results = append(results, GoogleSearchResult{
			Id:      item.CacheId,
			Link:    item.Link,
			Title:   item.Title,
			Snippet: item.Snippet,
		})
	}

	return results, nil
}

func (srv *googleSearchFunction) Stringify(results []GoogleSearchResult) (string, error) {
	jsonBytes, err := json.Marshal(results)
	if err != nil {
		return "", err
	}

	// Convert to string and remove spaces after colons
	jsonString := string(jsonBytes)
	jsonString = strings.ReplaceAll(jsonString, "\": ", "\":")

	return jsonString, nil
}
