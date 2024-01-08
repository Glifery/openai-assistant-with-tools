package web_crawler

import (
	"encoding/json"
	"github.com/glifery/openai-assistant-with-tools/types"
	"io/ioutil"
	"net/http"
)

type WebCrawlerRequest struct {
	Url string `json:"url" func:"Website URL that the user wants to crawl. It is a required field.,required"`
}

type WebCrawlerResult struct {
	Url     string `json:"url"`
	Content string `json:"content"`
}

type webCrawlerFunction struct {
}

func NewWebCrawlerFunction() types.FunctionCalling {
	return &webCrawlerFunction{}
}

func (srv *webCrawlerFunction) GetFunctionName() string {
	return "web-crawler-function"
}

func (srv *webCrawlerFunction) GetFunctionDescription() string {
	return "This function parses website content. It gets website URL as an input then it opens the website and returns the HTML content."
}

func (srv *webCrawlerFunction) GetFunctionProperties() map[string]map[string]interface{} {
	return map[string]map[string]interface{}{
		"url": {
			"type":        "string",
			"description": "Website URL that the user wants to crawl. It is a required field.",
		},
	}
}

func (srv *webCrawlerFunction) GetFunctionRequiredProperties() []string {
	return []string{"url"}
}

func (srv *webCrawlerFunction) Execute(input string) (output any, err error) {
	var request WebCrawlerRequest
	err = json.Unmarshal([]byte(input), &request)
	if err != nil {
		return nil, err
	}

	htmlData, err := srv.crawl(request.Url)
	if err != nil {
		return nil, err
	}

	return htmlData, nil
	//return WebCrawlerResult{
	//	Url:     request.Url,
	//	Content: htmlData,
	//}, nil
}

func (srv *webCrawlerFunction) crawl(url string) (result string, err error) {
	// Send HTTP GET request to the specified URL
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	// Read the HTML content from the response body
	htmlData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	// Print the HTML content
	return string(htmlData), nil
}
