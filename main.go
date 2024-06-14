package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

type Candidate struct {
	Index         int     `json:"Index"`
	Content       Content `json:"Content"`
	FinishReason  int     `json:"FinishReason"`
	SafetyRatings []struct {
		Category    int  `json:"Category"`
		Probability int  `json:"Probability"`
		Blocked     bool `json:"Blocked"`
	} `json:"SafetyRatings"`
	CitationMetadata interface{} `json:"CitationMetadata"`
	TokenCount       int         `json:"TokenCount"`
}

type Response struct {
	Candidates     []Candidate `json:"Candidates"`
	PromptFeedback interface{} `json:"PromptFeedback"`
	UsageMetadata  struct {
		PromptTokenCount     int `json:"PromptTokenCount"`
		CandidatesTokenCount int `json:"CandidatesTokenCount"`
		TotalTokenCount      int `json:"TotalTokenCount"`
	} `json:"UsageMetadata"`
}

type Content struct {
	Parts []string `json:"Parts"`
	Role  string   `json:"Role"`
}
type Candidates struct {
	Content *Content `json:"Content"`
}
type ContentResponse struct {
	Candidates *[]Candidates `json:"Candidates"`
}
type BingAnswer struct {
	Type            string   `json:"_type"`
	QueryContext    struct{} `json:"queryContext"`
	WebPages        WebPages `json:"webPages"`
	RelatedSearches struct{} `json:"relatedSearches"`
	RankingResponse struct{} `json:"rankingResponse"`
}

type WebPages struct {
	WebSearchURL          string   `json:"webSearchUrl"`
	TotalEstimatedMatches int      `json:"totalEstimatedMatches"`
	Value                 []Result `json:"value"`
}

type Result struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	URL              string    `json:"url"`
	IsFamilyFriendly bool      `json:"isFamilyFriendly"`
	DisplayURL       string    `json:"displayUrl"`
	Snippet          string    `json:"snippet"`
	DateLastCrawled  time.Time `json:"dateLastCrawled"`
	SearchTags       []struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	} `json:"searchTags,omitempty"`
	About []struct {
		Name string `json:"name"`
	} `json:"about,omitempty"`
}

func keywords(a string) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env key")
	}
	apiKey := os.Getenv("API_KEY")

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-pro")

	prompt := []genai.Part{
		genai.Text(a),
	}

	resp, err := model.GenerateContent(ctx, prompt...)
	if err != nil {
		log.Fatal(err)
	}

	for _, candidate := range resp.Candidates {
		fmt.Println(candidate.Content.Parts[len(candidate.Content.Parts)-1])
	}
}
func bingSearch(endpoint, token, query string) (*BingAnswer, error) {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	param := req.URL.Query()
	param.Add("q", query)
	req.URL.RawQuery = param.Encode()

	req.Header.Add("Ocp-Apim-Subscription-Key", token)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	var ans BingAnswer
	if err := json.Unmarshal(body, &ans); err != nil {
		return nil, fmt.Errorf("unmarshaling response: %w", err)
	}

	return &ans, nil
}

func main() {

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter a string: ")
	scanner.Scan()
	s := scanner.Text()
	var a string = "1.'Query: What is the outlook for Infosys' revenue growth in the next quarter?Searchterms: 'after:2024-06-13 Infosys revenue growth forecast' 2. Query: How has the conflict in Ukraine impacted U.S. steel imports in 2024? Search terms:'after:2024 Ukraine steel imports US''steel prices US after:2024-02-24'3. Query: What are the credit rating changes for Adani Group companies in the last 6 months? Searchterms: 'after:2023-12-13 Adani Group credit rating'4. Query: How has the RBI's monetary policy impacted interest rates in India over the past year? Search terms:'after:2023-06-13 RBI monetary policy interest rates India''RBI repo rate after:2023-06-13'5. Query: What are analysts' expectations for the future performance of the Indian stock market? Searchterms: 'India stock market outlook 2024 analysts'6. Query: How has the performance of mutual funds investing in the IT sector changed in 2024? Searchterms: 'after:2024 IT mutual fund performance'7. Query: What are the factors driving up gold prices globally in recent months?Search terms:'after:2024 gold price increase factors''global economic factors affecting gold price after:2024'8. Query: How has the merger between HDFC Bank and Bank of Baroda affected their stock prices? Search terms:'HDFC Bank Bank of Baroda merger after:2023''after:2023 HDFC Bank stock price''after:2023 Bank of Baroda stock price'9. Query: What is the impact of inflation on consumer spending in India? Search terms:'inflation and consumer spending India after:2024' 'RBI inflation report after:2024'10. Query:How has the cryptocurrency market performed in the first half of 2024? Searchterms: 'cryptocurrency market performance after:2024-01-01' 11.Query: What is Hdfc bank saying about RELIANCE Searchterms: site:hdfcbank.com “Reliance Industries”. These are relevant search terms, which are basically keywords for the query to use for the respective mentioned queries.  Now provide Search term/s for the Query:"
	var b string = a + s + ". Just provide the Search term/s for this statement and dont write anything else. dont use even 'Search terms=', just the search terms for the query, also relevant site is mandatory in each of your answer search terms "
	file, err := os.CreateTemp("", "output.txt")
	if err != nil {
		fmt.Println("Error creating temporary file:", err)
		return
	}
	defer os.Remove(file.Name())
	old := os.Stdout
	os.Stdout = file

	keywords(b)

	os.Stdout = old

	file.Close()

	// content, err := os.ReadFile(file.Name())
	// if err != nil {
	// 	fmt.Println("Error reading file content:", err)
	// 	return
	// }

	// keywordsString := string(content)

	// const (
	// 	endpoint = "https://api.bing.microsoft.com/v7.0/search"
	// 	token    = "e635fcdf348e4a868154deb206dc0740"
	// )
	// var searchTerm = keywordsString
	// ans, err := bingSearch(endpoint, token, searchTerm)
	// if err != nil {
	// 	log.Fatalf("Failed to get search results: %v", err)
	// }

	var final_string string
	// for _, result := range ans.WebPages.Value {
	// 	fmt.Printf("Name: %s\nURL: %s\nDescription: %s\n\n", result.Name, result.URL, result.Snippet)
	// 	final_string = final_string + "Name: " + result.Name + "\n" + "Description: " + result.Snippet + "\n"
	// }
	final_string = b
	keywords(final_string)
}
