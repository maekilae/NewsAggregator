package classifier

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"newsaggregator/internal/article"
	"os"
	"time"
)

const RESPONSE_FORMAT = `
{
  "type": "json_schema",
  "json_schema": {
    "name": "classification_response",
    "strict": true,
    "schema": {
      "type": "object",
      "title": "ListResponse",
      "properties": {
        "items": {
          "type": "array",
          "description": "List of classified items",
          "items": {
            "type": "object",
            "properties": {
              "index": {
                "type": "number",
                "description": "Item Index"
              },
              "class": {
                "type": "string",
                "description": "Item Classification"
              }
            },
            "required": [
              "index",
              "class"
            ],
            "additionalProperties": false
          }
        }
      },
      "required": [
        "items"
      ],
      "additionalProperties": false
    }
  }
}
`
const SYSTEM_MSG = `
### Role
You are a precise news classifier. Your task is to process news items and assign them categories from a strict taxonomy.

### Instructions
1. **Classification:** Assign the most relevant category from the "Allowed Classifications" list.
2. **Dual Tagging:** If an item fits two categories equally, combine them with a forward slash (e.g., "Political/Geopolitics"). Limit to two categories.
3. **Strict Adherence:** Use ONLY the categories provided. Do not create new tags.
4. **Indexing:** Maintain the original order using the 'index' key (starting at 0 for the first news item provided).
5. **Output:** Return a JSON object containing an "items" array.

### Allowed Classifications
- General Summary, Entertainment, Culture, Political, Social Issues, Public Safety, Incident, Environmental, Investigative Journalism, Crime, Justice, Accident, Health, Human Interest, Consumer, Business, Local Government, Public Services, Geopolitics, Sports

### Constraints
- Prioritize the "Primary Driver" (e.g., a story about a "Criminal investigation" into a "Politician" is "Political/Crime").
- If a story is purely about a sports result, use "Sports".`

type Classifier struct {
	apiKey string
	apiUrl string
}

type requestData struct {
	Model          string          `json:"model"`
	Messages       []Message       `json:"messages"`
	ResponseFormat json.RawMessage `json:"response_format,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ClassificationResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// 2. Your custom structured output (The inner JSON)
type ClassificationData struct {
	Items []struct {
		Index int    `json:"index"`
		Class string `json:"class"`
	} `json:"items"`
}

func New(keyEnv string, apiUrl string) *Classifier {
	return &Classifier{
		apiKey: os.Getenv(keyEnv),
		apiUrl: apiUrl,
	}
}

func (c *Classifier) Classify(articles map[string]article.Article) (map[string]article.Article, error) {
	if c.apiKey == "" {
		return nil, errors.New("API key not set")
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	keys := make([]string, 0, len(articles))
	for k, article := range articles {
		encoder.Encode(article.Title)
		keys = append(keys, k)
	}
	reqData := requestData{
		Model: "mistral-small-latest",
		Messages: []Message{
			{Role: "system", Content: SYSTEM_MSG},
			{Role: "user", Content: buf.String()},
		},
		ResponseFormat: json.RawMessage(RESPONSE_FORMAT),
	}
	payload, err := json.Marshal(reqData)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("POST", c.apiUrl, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	client := &http.Client{
		Timeout: 10 * time.Second, // Always set a timeout!
	}
	dump, _ := httputil.DumpRequest(req, true)
	fmt.Println(string(dump))
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Step 1: Unmarshal the outer Mistral wrapper
	var mistralResp ClassificationResponse
	if err := json.Unmarshal(bodyBytes, &mistralResp); err != nil {
		return nil, err
	}

	// Safety check: Make sure we actually got a choice back
	if len(mistralResp.Choices) == 0 {
		return nil, errors.New("no choices")
	}

	// Extract the raw JSON string the AI generated
	rawAIContent := mistralResp.Choices[0].Message.Content

	// Step 2: Unmarshal the AI's JSON string into your custom struct
	var result ClassificationData
	if err := json.Unmarshal([]byte(rawAIContent), &result); err != nil {
		return nil, err
	}

	i := 0
	for _, k := range keys {
		art := articles[k]
		art.Tag = result.Items[i].Class
		articles[k] = art
		i++
		if len(result.Items) <= i {
			break
		}
	}

	return articles, nil
}
